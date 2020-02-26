/*
 * Copyright 2019 The original author or authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
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
	"github.com/Shopify/sarama"
	"github.com/projectriff/kafka-provisioner/pkg/provisioner/handler"
	client "github.com/projectriff/kafka-provisioner/pkg/provisioner/kafka"
	"log"
	"net/http"
	"os"
)

func main() {
	gateway := os.Getenv("GATEWAY")
	if gateway == "" {
		log.Fatal("Environment variable GATEWAY should contain the host and port of a liiklus gRPC endpoint")
	}
	broker := os.Getenv("BROKER")
	if broker == "" {
		log.Fatal("Environment variable BROKER should contain the host and port of a Kafka broker")
	}

	sarama.Logger = log.New(os.Stdout, "[Sarama] ", log.LstdFlags)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		handleProvisionRequest(broker, gateway, w, r)
	})
	_ = http.ListenAndServe(":8080", nil)
}

func handleProvisionRequest(broker, gateway string, writer http.ResponseWriter, request *http.Request) {
	kafkaClient, err := client.NewKafkaClient(broker)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		_, _ = fmt.Fprintf(os.Stderr, "Error connecting to Kafka broker %q: %v\n", broker, err)
		_, _ = fmt.Fprintf(writer, "Error connecting to Kafka broker %q: %v\n", broker, err)
		return
	}
	defer func() {
		if err := kafkaClient.Close(); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Error disconnecting from Kafka broker %q: %v\n", broker, err)
		}
	}()
	requestHandler := &handler.TopicCreationRequestHandler{KafkaClient: kafkaClient, Gateway: gateway, Writer: os.Stderr}
	requestHandler.GetHandlerFunc()(writer, request)
}
