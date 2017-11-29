/*
 * Copyright 2017 the original author or authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/Shopify/sarama"
	"github.com/bsm/sarama-cluster"

	"flag"
	dispatch "github.com/projectriff/function-sidecar/pkg/dispatcher"
	"github.com/projectriff/function-sidecar/pkg/dispatcher/grpc"
	"github.com/projectriff/function-sidecar/pkg/dispatcher/http"
	"github.com/projectriff/function-sidecar/pkg/dispatcher/pipes"
	"github.com/projectriff/function-sidecar/pkg/dispatcher/stdio"
	"github.com/projectriff/function-sidecar/pkg/wireformat"
	"io"
	"strings"
)

type stringSlice []string

func (sl *stringSlice) String() string {
	return fmt.Sprint(*sl)
}

func (sl *stringSlice) Set(value string) error {
	*sl = stringSlice(strings.Split(value, ","))
	return nil
}

var brokers stringSlice = []string{"localhost:9092"}
var inputs, outputs stringSlice
var group, protocol string

func init() {
	flag.Var(&brokers, "brokers", "location of the Kafka server(s) to connect to")
	flag.Var(&inputs, "inputs", "kafka topic(s) to listen to, as input for the function")
	flag.Var(&outputs, "outputs", "kafka topic(s) to write to with results from the function")
	flag.StringVar(&group, "group", "", "kafka consumer group to act as")
	flag.StringVar(&protocol, "protocol", "", "dispatcher protocol to use. One of [http, grpc, stdio]")
}

const CorrelationId = "correlationId"

// PropagatedHeaders is the set of header names that will be copied from the incoming message
// to the outgoing message
var PropagatedHeaders = []string{CorrelationId}

func main() {

	flag.Parse()

	if len(inputs) > 1 {
		log.Fatalf("Only 1 input supported for now. See https://github.com/projectriff/riff/issues/184. Provided %v\n", inputs)
	} else if len(outputs) > 1 {
		log.Fatalf("Only 1 output supported for now. See https://github.com/projectriff/riff/issues/184. Provided %v\n", outputs)
	}

	input := inputs[0]
	var output string
	if len(outputs) > 0 {
		output = outputs[0]
	} else {
		output = ""
	}

	log.Printf("Sidecar for function '%v' (%v->%v) using %v dispatcher starting\n", group, input, output, protocol)

	var producer sarama.AsyncProducer
	var err error
	if output != "" {
		producer, err = sarama.NewAsyncProducer(brokers, nil)
		if err != nil {
			panic(err)
		}
		defer producer.Close()
	}

	consumerConfig := makeConsumerConfig()
	consumerConfig.Consumer.Offsets.Initial = sarama.OffsetOldest

	consumer, err := cluster.NewConsumer(brokers, group, []string{input}, consumerConfig)
	if err != nil {
		panic(err)
	}
	defer consumer.Close()
	if consumerConfig.Consumer.Return.Errors {
		go consumeErrors(consumer)
	}
	if consumerConfig.Group.Return.Notifications {
		go consumeNotifications(consumer)
	}

	dispatcher, err := createDispatcher(protocol)
	if err != nil {
		panic(err)
	}
	switch d := dispatcher.(type) {
	case io.Closer:
		log.Print("Requesting close()")
		defer d.Close()
	}

	// trap SIGINT, SIGTERM, and SIGKILL to trigger a shutdown.
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM, os.Kill)

	go func() {
		for {
			select {
			// Incoming message
			case msg, open := <-consumer.Messages():
				if open {
					messageIn, err := wireformat.FromKafka(msg)
					log.Printf(">>> %s\n", messageIn)
					if err != nil {
						log.Printf("Error receiving message from Kafka: %v", err)
						break
					}
					dispatcher.Input() <- messageIn
					consumer.MarkOffset(msg, "") // mark message as processed
				} else {
					// Kafka closed
					log.Print("Exiting Kafka Consumer loop")
					return
				}
			}
		}
	}()

	go func() {
		for {
			select {
			// Result message
			case resultMsg, open := <-dispatcher.Output(): // Make sure to drain channel even if output==""
				if open {
					if output != "" {
						log.Printf("<<< %s\n", resultMsg)
						producerMessage, err := wireformat.ToKafka(resultMsg)
						if err != nil {
							log.Printf("Error encoding message: %v", err)
							break
						}
						producerMessage.Topic = output
						select {
						case producer.Input() <- producerMessage:
						}
					} else {
						log.Printf("=== Not sending function return value as function did not provide an output channel. Raw result = %s\n", resultMsg)
					}
				} else {
					log.Print("Exiting Kafka Producer loop")
					return
				}
			}
		}
	}()

	// consume messages, watch signals
	for {
		select {
		// Request for shutdown
		case <-signals:
			close(dispatcher.Input())
			return
		}
	}
}

// copy headers from incomingHeaders that need to be propagated into resultHeaders
func propagateHeaders(incomingHeaders dispatch.Headers, resultHeaders dispatch.Headers) dispatch.Headers {
	result := make(dispatch.Headers)
	if resultHeaders != nil {
		for k, v := range resultHeaders {
			result[k] = v
		}
	}
	for _, h := range PropagatedHeaders {
		if value, ok := incomingHeaders[h]; ok {
			result[h] = value
		}
	}
	return result
}

func createDispatcher(protocol string) (dispatch.Dispatcher, error) {
	switch protocol {
	case "http":
		return dispatch.NewWrapper(http.NewHttpDispatcher())
	case "pipes":
		return pipes.NewPipesDispatcher()
	case "stdio":
		d, err := stdio.NewStdioDispatcher()
		if err != nil {
			return nil, err
		}
		return dispatch.NewWrapper(d)
	case "grpc":
		return dispatch.NewWrapper(grpc.NewGrpcDispatcher())
	default:
		panic("Unsupported Dispatcher " + protocol)
	}
}

func consumeNotifications(consumer *cluster.Consumer) {
	for ntf := range consumer.Notifications() {
		log.Printf("Rebalanced: %+v\n", ntf)
	}
}

func consumeErrors(consumer *cluster.Consumer) {
	for err := range consumer.Errors() {
		log.Printf("Error: %s\n", err.Error())
	}
}

func makeConsumerConfig() *cluster.Config {
	consumerConfig := cluster.NewConfig()
	consumerConfig.Consumer.Return.Errors = true
	consumerConfig.Group.Return.Notifications = true
	return consumerConfig
}
