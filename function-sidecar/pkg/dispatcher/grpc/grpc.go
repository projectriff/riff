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

package grpc

import (
	"google.golang.org/grpc"

	"fmt"
	"log"
	"time"

	"github.com/projectriff/function-sidecar/pkg/dispatcher"
	"github.com/projectriff/function-sidecar/pkg/dispatcher/grpc/function"
	"github.com/projectriff/message-transport/pkg/message"
	"golang.org/x/net/context"
	"io"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/codes"
)

type grpcDispatcher struct {
	stream function.MessageFunction_CallClient
	input  chan message.Message
	output chan message.Message
}

func (this *grpcDispatcher) Input() chan<- message.Message {
	return this.input
}

func (this *grpcDispatcher) Output() <-chan message.Message {
	return this.output
}

func (this *grpcDispatcher) handleIncoming() {
	for {
		select {
		case in, open := <-this.input:
			if open {
				grpcMessage := toGRPC(in)
				err := this.stream.Send(grpcMessage)
				if err != nil {
					if streamClosureDiagnosed(err) {
						return
					}

					log.Printf("Error sending message to function: %v", err)
				}
			} else {
				close(this.output)
				log.Print("Shutting down gRPC dispatcher")
				return
			}
		}
	}
}

func (this *grpcDispatcher) handleOutgoing() {
	for {
		reply, err := this.stream.Recv()
		if err != nil {
			if streamClosureDiagnosed(err) {
				return
			}

			log.Printf("Error receiving message from function: %v", err)
			continue
		}
		message := toDispatcher(reply)
		this.output <- message
	}
}

func streamClosureDiagnosed(err error) bool {
	if err == io.EOF {
		log.Println("Stream to function has closed")
		return true
	}

	if sErr, ok := status.FromError(err); ok && sErr.Code() == codes.Unavailable {
		// See https://github.com/grpc/grpc/blob/master/doc/statuscodes.md
		log.Printf("Stream to function is closing: %v", err)
		return true
	}

	return false
}

func NewGrpcDispatcher(port int) (dispatcher.Dispatcher, error) {
	ctx, _ := context.WithTimeout(context.Background(), 60*time.Second)
	conn, err := grpc.DialContext(ctx, fmt.Sprintf("localhost:%v", port), grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		return nil, err
	}

	fnStream, err := function.NewMessageFunctionClient(conn).Call(context.Background())
	if err != nil {
		return nil, err
	}

	result := &grpcDispatcher{fnStream, make(chan message.Message, 100), make(chan message.Message, 100)}
	go result.handleIncoming()
	go result.handleOutgoing()

	return result, nil
}

func toGRPC(message message.Message) *function.Message {
	grpcHeaders := make(map[string]*function.Message_HeaderValue, len(message.Headers()))
	for k, vv := range message.Headers() {
		values := function.Message_HeaderValue{}
		grpcHeaders[k] = &values
		for _, v := range vv {
			values.Values = append(values.Values, v)
		}
	}
	result := function.Message{Payload: message.Payload(), Headers: grpcHeaders}

	return &result
}

func toDispatcher(grpc *function.Message) message.Message {
	dHeaders := make(map[string][]string, len(grpc.Headers))
	for k, pv := range grpc.Headers {
		dHeaders[k] = pv.Values
	}
	return message.NewMessage(grpc.Payload, dHeaders)
}
