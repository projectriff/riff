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
	"context"
	"fmt"
	"gopkg.in/Shopify/sarama.v1"
	"github.com/bsm/sarama-cluster"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"
	"github.com/satori/go.uuid"
	"github.com/sk8sio/http-gateway/pkg/message"
	"syscall"
)

// Function messageHandler creates an http handler that posts the http body as a message to Kafka, replying
// immediately with a successful http response
func messageHandler(producer sarama.AsyncProducer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		topic := r.URL.Path[len("/messages/"):]
		b, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		scsMessage := message.Message{Payload: b, Headers: nil}

		bytesOut, err := message.EncodeMessage(scsMessage)
		if err != nil {
			log.Printf("Error encoding message: %v", err)
			return
		}
		kafkaMsg := &sarama.ProducerMessage{Topic: topic, Value: sarama.ByteEncoder(bytesOut)}

		select {
		case producer.Input() <- kafkaMsg:
			w.Write([]byte("message published to topic: " + topic + "\n"))
		}
	}
}

// Function replyHandler creates an http handler that posts the http body as a message to Kafka, then waits
// for a message on a go channel it creates for a reply (this is expected to be set by the main thread) and sends
// that as an http response.
func replyHandler(producer sarama.AsyncProducer, replies map[string]chan message.Message) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		topic := r.URL.Path[len("/requests/"):]
		b, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		correlationId := uuid.NewV4().String()
		replyChan := make(chan message.Message)
		replies[correlationId] = replyChan
		scsMessage := message.Message{Payload: b, Headers: map[string]interface{}{"correlationId": correlationId}}

		bytesOut, err := message.EncodeMessage(scsMessage)
		if err != nil {
			log.Printf("Error encoding message: %v", err)
			return
		}
		kafkaMsg := &sarama.ProducerMessage{Topic: topic, Value: sarama.ByteEncoder(bytesOut)}

		select {
		case producer.Input() <- kafkaMsg:
			select {
			case reply := <-replyChan:
				delete(replies, correlationId)
				p := reply.Payload
				w.Write(p.([]byte)) // TODO equivalent of Spring's HttpMessageConverter handling
			case <- time.After(time.Second * 60):
				delete(replies, correlationId)
				w.WriteHeader(404)
			}
		}
	}
}

func healthHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"status":"UP"}`))
	}
}

func startHttpServer(producer sarama.AsyncProducer, replies map[string]chan message.Message) *http.Server {
	srv := &http.Server{Addr: ":8080"}

	http.HandleFunc("/messages/", messageHandler(producer))
	http.HandleFunc("/requests/", replyHandler(producer, replies))
	http.HandleFunc("/application/status", healthHandler())

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Printf("Httpserver: ListenAndServe() error: %s", err)
		}
	}()

	log.Printf("Listening on %v", srv.Addr)
	return srv
}

func main() {
	// Trap signals to trigger a proper shutdown.
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM, os.Kill)

	// Key is correlationId, value is channel used to pass message received from main Kafka consumer loop
	replies := make(map[string]chan message.Message)

	brokers := []string{os.Getenv("SPRING_CLOUD_STREAM_KAFKA_BINDER_BROKERS")}
	producer, err := sarama.NewAsyncProducer(brokers, nil)
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := producer.Close(); err != nil {
			log.Fatalln(err)
		}
	}()

	consumerConfig := makeConsumerConfig()
	consumer, err := cluster.NewConsumer(brokers, "gateway", []string{"replies"}, consumerConfig)
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



	srv := startHttpServer(producer, replies)

MainLoop:
	for {
		select {
		case <-signals:
			fmt.Println("Shutting Down...")
			timeout, c := context.WithTimeout(context.Background(), 1*time.Second)
			defer c()
			if err := srv.Shutdown(timeout); err != nil {
				panic(err) // failure/timeout shutting down the server gracefully
			}
			break MainLoop
		case msg, ok := <-consumer.Messages():
			if ok {
				messageWithHeaders, err := message.ExtractMessage(msg.Value)
				if err != nil {
					log.Println("Failed to extract message ", err)
					break
				}
				correlationId, ok := messageWithHeaders.Headers["correlationId"].(string)
				if ok {
					c := replies[correlationId]
					if c != nil {
						fmt.Printf("Sending %v\n", messageWithHeaders)
						c <- messageWithHeaders
						consumer.MarkOffset(msg, "") // mark message as processed
					} else {
						log.Printf("Did not find communication channel for correlationId %v. Timed out?", correlationId)
						consumer.MarkOffset(msg, "") // mark message as processed
					}
				}
			}

		case err := <-producer.Errors():
			log.Println("Failed to produce kafka message ", err)
		}
	}
}

func makeConsumerConfig() *cluster.Config {
	consumerConfig := cluster.NewConfig()
	consumerConfig.Consumer.Return.Errors = true
	consumerConfig.Group.Return.Notifications = true
	return consumerConfig
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
