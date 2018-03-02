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
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/bsm/sarama-cluster"
	"github.com/projectriff/riff/http-gateway/pkg/server"
	"github.com/projectriff/riff/message-transport/pkg/transport/kafka"
	"github.com/projectriff/riff/message-transport/pkg/transport/metrics/kafka_over_kafka"
	"github.com/satori/go.uuid"
)

func main() {

	brokers := brokers()
	producerId := uuid.NewV4().String()
	producer, err := kafka_over_kafka.NewMetricsEmittingProducer(brokers, producerId)
	if err != nil {
		panic(err)
	}
	defer producer.Close()

	consumer, err := kafka.NewConsumer(brokers, "gateway", []string{"replies"}, cluster.NewConfig())
	if err != nil {
		panic(err)
	}
	defer consumer.Close()

	gw := server.New(8080, producer, consumer, 60*time.Second)

	done := make(chan struct{})
	gw.Run(done)

	// Wait for shutdown
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM, os.Kill)
	<-signals
	log.Println("Shutting Down...")
	close(done)

}

func brokers() []string {
	return strings.Split(os.Getenv("KAFKA_BROKERS"), ",")
}
