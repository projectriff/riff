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
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/Shopify/sarama"
	"github.com/bsm/sarama-cluster"
	"github.com/projectriff/function-sidecar/pkg/dispatcher"
	"github.com/projectriff/function-sidecar/pkg/wireformat"
	"github.com/satori/go.uuid"
)

const ContentType = "Content-Type"
const Accept = "Accept"
const CorrelationId = "correlationId"

var incomingHeadersToPropagate = [...]string{ContentType, Accept}
var outgoingHeadersToPropagate = [...]string{ContentType}

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
		msg := dispatcher.NewMessage(b, make(map[string][]string))
		propagateIncomingHeaders(r, msg)

		kafkaMsg, err := wireformat.ToKafka(msg)
		kafkaMsg.Topic = topic
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		producer.Input() <- kafkaMsg
		w.Write([]byte("message published to topic: " + topic + "\n"))
	}
}

// Function replyHandler creates an http handler that posts the http body as a message to Kafka, then waits
// for a message on a go channel it creates for a reply (this is expected to be set by the main thread) and sends
// that as an http response.
func replyHandler(producer sarama.AsyncProducer, replies *repliesMap) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		topic := r.URL.Path[len("/requests/"):]
		b, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		correlationId := uuid.NewV4().String()
		replyChan := make(chan dispatcher.Message)
		replies.put(correlationId, replyChan)

		msg := dispatcher.NewMessage(b, make(map[string][]string))
		propagateIncomingHeaders(r, msg)
		msg.Headers()[CorrelationId] = []string{correlationId}

		kafkaMsg, err := wireformat.ToKafka(msg)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		kafkaMsg.Topic = topic

		producer.Input() <- kafkaMsg

		select {
		case reply := <-replyChan:
			replies.delete(correlationId)
			propagateOutgoingHeaders(reply, w)
			w.Write(reply.Payload())
		case <-time.After(time.Second * 60):
			replies.delete(correlationId)
			w.WriteHeader(404)
		}
	}
}
func healthHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"status":"UP"}`))
	}
}

func startHttpServer(producer sarama.AsyncProducer, replies *repliesMap) *http.Server {
	srv := &http.Server{Addr: ":8080"}

	http.HandleFunc("/messages/", messageHandler(producer))
	http.HandleFunc("/requests/", replyHandler(producer, replies))
	http.HandleFunc("/application/status", healthHandler())

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			panic(err)
		}
	}()

	log.Printf("Listening on %v", srv.Addr)
	return srv
}

func propagateIncomingHeaders(request *http.Request, message dispatcher.Message) {
	for _, h := range incomingHeadersToPropagate {
		if vs, ok := request.Header[h]; ok {
			(message.Headers())[h] = vs
		}
	}
}

func propagateOutgoingHeaders(message dispatcher.Message, response http.ResponseWriter) {
	for _, h := range outgoingHeadersToPropagate {
		if vs, ok := message.Headers()[h]; ok {
			response.Header()[h] = vs
		}
	}
}

func main() {
	// Trap signals to trigger a proper shutdown.
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM, os.Kill)

	// Key is correlationId, value is channel used to pass message received from main Kafka consumer loop
	replies := newRepliesMap()

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
			log.Println("Shutting Down...")
			timeout, c := context.WithTimeout(context.Background(), 1*time.Second)
			defer c()
			if err := srv.Shutdown(timeout); err != nil {
				panic(err) // failure/timeout shutting down the server gracefully
			}
			break MainLoop
		case msg, ok := <-consumer.Messages():
			if ok {
				messageWithHeaders, err := wireformat.FromKafka(msg)
				if err != nil {
					log.Println("Failed to extract message ", err)
					break
				}
				correlationId, ok := messageWithHeaders.Headers()[CorrelationId]
				if ok {
					c := replies.get(correlationId[0])
					if c != nil {
						log.Printf("Sending %v\n", messageWithHeaders)
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

// Type repliesMap implements a concurrent safe map of channels to send replies to, keyed by message correlationIds
type repliesMap struct {
	m    map[string]chan<- dispatcher.Message
	lock sync.RWMutex
}

func (replies *repliesMap) delete(key string) {
	replies.lock.Lock()
	defer replies.lock.Unlock()
	delete(replies.m, key)
}

func (replies *repliesMap) get(key string) chan<- dispatcher.Message {
	replies.lock.RLock()
	defer replies.lock.RUnlock()
	return replies.m[key]
}

func (replies *repliesMap) put(key string, value chan<- dispatcher.Message) {
	replies.lock.Lock()
	defer replies.lock.Unlock()
	replies.m[key] = value
}

func newRepliesMap() *repliesMap {
	return &repliesMap{make(map[string]chan<- dispatcher.Message), sync.RWMutex{}}
}
