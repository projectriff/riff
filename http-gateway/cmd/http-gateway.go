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
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"
	"github.com/sk8sio/http-gateway/pkg/message"
)

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

func healthHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"status":"UP"}`))
	}
}


func startHttpServer(producer sarama.AsyncProducer) *http.Server {
	srv := &http.Server{Addr: ":8080"}

	http.HandleFunc("/messages/", messageHandler(producer))
	http.HandleFunc("/health", healthHandler())

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Printf("Httpserver: ListenAndServe() error: %s", err)
		}
	}()

	log.Printf("Listening on %v", srv.Addr)
	return srv
}

func main() {
	// Trap SIGINT to trigger a shutdown.
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)

	producer, err := sarama.NewAsyncProducer([]string{os.Getenv("SPRING_CLOUD_STREAM_KAFKA_BINDER_BROKERS")}, nil)
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := producer.Close(); err != nil {
			log.Fatalln(err)
		}
	}()

	srv := startHttpServer(producer)


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
		case err := <-producer.Errors():
			log.Println("Failed to produce kafka message ", err)
		}
	}
}
