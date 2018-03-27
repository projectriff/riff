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

package server

import (
	"time"
	"log"
	"net/http"
	"fmt"
	"github.com/projectriff/riff/message-transport/pkg/transport"
	"io"
	"context"
	"github.com/projectriff/riff/message-transport/pkg/message"
)

type Gateway interface {
	Run(stopCh <-chan struct{})
}

type gateway struct {
	httpServer       *http.Server
	consumer         transport.Consumer
	consumerMessages chan message.Message
	producer         transport.Producer
	replies          *repliesMap
	timeout          time.Duration
}

func (g *gateway) Run(stop <-chan struct{}) {
	go func() {
		log.Printf("Listening on %v", g.httpServer.Addr)
		if err := g.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			panic(err)
		}
	}()

	go g.repliesLoop(stop)
}

func (g *gateway) consumeRepliesLoop() {
	for {
		msg, _, err := g.consumer.Receive()
		if err != nil {
			break
		}
		g.consumerMessages <- msg
	}
}

func (g *gateway) repliesLoop(stop <-chan struct{}) {
	go g.consumeRepliesLoop()
	consumerMessages := g.consumerMessages
	producerErrors := g.producer.Errors()
	for {
		select {
		case msg, ok := <-consumerMessages:
			if ok {
				correlationId, ok := msg.Headers()[CorrelationId]
				if ok {
					c := g.replies.Get(correlationId[0])
					if c != nil {
						log.Printf("Sending reply %v\n", msg)
						c <- msg
					} else {
						log.Printf("Did not find communication channel for correlationId %v. Timed out?", correlationId)
					}
				}
			} else {
				break
			}
		case err := <-producerErrors:
			log.Println("Failed to send message ", err)
		case <-stop:
			if pCloseable, ok := g.producer.(io.Closer); ok {
				pCloseable.Close()
			}
			if cCloseable, ok := g.consumer.(io.Closer); ok {
				cCloseable.Close()
			}
			timeout, c := context.WithTimeout(context.Background(), 1*time.Second)
			defer c()
			if err := g.httpServer.Shutdown(timeout); err != nil {
				panic(err) // failure/timeout shutting down the server gracefully
			}
		}

	}
}

func New(port int, producer transport.Producer, consumer transport.Consumer, timeout time.Duration) *gateway {
	mux := http.NewServeMux()
	httpServer := &http.Server{Addr: fmt.Sprintf(":%v", port),
		Handler: mux,
	}
	consumerMessages := make(chan message.Message)
	g := gateway{httpServer: httpServer,
		producer: producer,
		consumer: consumer,
		consumerMessages: consumerMessages,
		replies: newRepliesMap(),
		timeout: timeout,
	}
	mux.HandleFunc(messagePath, g.messagesHandler)
	mux.HandleFunc(requestPath, g.requestsHandler)
	mux.HandleFunc("/application/status", healthHandler)

	return &g
}
