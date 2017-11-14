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

	"github.com/bsm/sarama-cluster"
	"encoding/json"

	"gopkg.in/Shopify/sarama.v1"

	"github.com/sk8sio/function-sidecar/pkg/dispatcher/http"
	"github.com/sk8sio/function-sidecar/pkg/dispatcher/stdio"
	"github.com/sk8sio/function-sidecar/pkg/dispatcher"
	"github.com/sk8sio/function-sidecar/pkg/message"
	"github.com/sk8sio/function-sidecar/pkg/dispatcher/grpc"
)

func main() {

	var saj map[string]interface{}
	err := json.Unmarshal([]byte(os.Getenv("SPRING_APPLICATION_JSON")), &saj)
	if err != nil {
		panic(err)
	}

	consumerConfig := makeConsumerConfig()
	consumerConfig.Consumer.Offsets.Initial = sarama.OffsetOldest

	brokers := []string{saj["spring.cloud.stream.kafka.binder.brokers"].(string)}
	input := saj["spring.cloud.stream.bindings.input.destination"].(string)
	output := saj["spring.cloud.stream.bindings.output.destination"]
	group := saj["spring.cloud.stream.bindings.input.group"].(string)
	protocol := saj["spring.profiles.active"].(string)

	fmt.Printf("Sidecar for function '%v' (%v->%v) using %v dispatcher starting\n", group, input, output, protocol)

	dispatcher := createDispatcher(protocol)

	var producer sarama.AsyncProducer
	if output != nil {
		producer, err = sarama.NewAsyncProducer(brokers, nil)
		if err != nil {
			panic(err)
		}
		defer producer.Close()
	}

	consumer, err := cluster.NewConsumer(brokers, group, []string{input}, consumerConfig)
	if err != nil {
		panic(err)
	}
	defer consumer.Close()

	// trap SIGINT to trigger a shutdown.
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)

	if consumerConfig.Consumer.Return.Errors {
		go consumeErrors(consumer)
	}
	if consumerConfig.Group.Return.Notifications {
		go consumeNotifications(consumer)
	}

	// consume messages, watch signals
	for {
		select {
		case msg, ok := <-consumer.Messages():
			if ok {
				messageIn, err := message.ExtractMessage(msg.Value)
				fmt.Fprintf(os.Stdout, "<<< %s\n", messageIn)
				if err != nil {
					log.Printf("Error receiving message from Kafka: %v", err)
					break
				}
				strPayload := string(messageIn.Payload.([]byte))
				dispatched, err := dispatcher.Dispatch(strPayload)
				if err != nil {
					log.Printf("Error dispatching message: %v", err)
					break
				}
				if output != nil {
					messageOut := message.Message{Payload: []byte(dispatched.(string)), Headers: messageIn.Headers}
					bytesOut, err := message.EncodeMessage(messageOut)
					fmt.Fprintf(os.Stdout, ">>> %s\n", messageOut)
					if err != nil {
						log.Printf("Error encoding message: %v", err)
						break
					}
					outMessage := &sarama.ProducerMessage{Topic: output.(string), Value: sarama.ByteEncoder(bytesOut)}
					producer.Input() <- outMessage
				} else {
					fmt.Fprintf(os.Stdout, "=== Not sending function return value as function did not provide an output channel. Raw result = %s\n", dispatched)
				}
				consumer.MarkOffset(msg, "") // mark message as processed
			}
		case <-signals:
			return
		}
	}
}
func createDispatcher(protocol string) dispatcher.Dispatcher {
	switch protocol {
	case "http":
		return http.NewHttpDispatcher()
	case "stdio":
		return stdio.NewStdioDispatcher()
	case "grpc":
		return grpc.NewGrpcDispatcher()
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
