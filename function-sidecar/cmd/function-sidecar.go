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

	"flag"
	"io"
	"strings"

	"github.com/Shopify/sarama"
	"github.com/bsm/sarama-cluster"
	"github.com/projectriff/riff/function-sidecar/pkg/carrier"
	dispatch "github.com/projectriff/riff/function-sidecar/pkg/dispatcher"
	"github.com/projectriff/riff/function-sidecar/pkg/dispatcher/grpc"
	"github.com/projectriff/riff/function-sidecar/pkg/dispatcher/http"
	"github.com/projectriff/riff/message-transport/pkg/transport"
	"github.com/projectriff/riff/message-transport/pkg/transport/metrics/kafka_over_kafka"
	"github.com/satori/go.uuid"
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
	flag.StringVar(&protocol, "protocol", "", "dispatcher protocol to use. One of [http, grpc]")
}

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

	var producer transport.Producer
	var err error
	if output != "" {
		producer, err = kafka_over_kafka.NewMetricsEmittingProducer(brokers, uuid.NewV4().String())
		if err != nil {
			panic(err)
		}

		if prod, ok := producer.(io.Closer); ok {
			defer prod.Close()
		}
	}

	consumerConfig := makeConsumerConfig()
	consumerConfig.Consumer.Offsets.Initial = sarama.OffsetOldest

	pod := uuid.NewV4().String()

	consumer, err := kafka_over_kafka.NewMetricsEmittingConsumer(brokers, group, pod, []string{input}, consumerConfig)
	if err != nil {
		panic(err)
	}
	defer consumer.Close()

	dispatcher, err := createDispatcher(protocol)
	if err != nil {
		panic(err)
	}
	switch d := dispatcher.(type) {
	case io.Closer:
		log.Print("Requesting close()")
		defer d.Close()
	}

	carrier.Run(consumer, producer, dispatcher, output)

	// trap SIGINT and SIGTERM to trigger a shutdown.
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
	<-signals
	log.Println(("Shutting Down..."))
}

func createDispatcher(protocol string) (dispatch.Dispatcher, error) {
	switch protocol {
	case "http":
		return dispatch.NewWrapper(http.NewHttpDispatcher())
	case "grpc":
		return grpc.NewGrpcDispatcher(10382)
	default:
		panic("Unsupported Dispatcher " + protocol)
	}
}

func makeConsumerConfig() *cluster.Config {
	consumerConfig := cluster.NewConfig()
	consumerConfig.Consumer.Return.Errors = true
	consumerConfig.Group.Return.Notifications = true
	return consumerConfig
}
