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

	function "github.com/projectriff/function-sidecar/pkg/dispatcher/grpc/function"
	fntypes "github.com/projectriff/function-sidecar/pkg/dispatcher/grpc/fntypes"
	"github.com/projectriff/function-sidecar/pkg/dispatcher"
	"log"
	"golang.org/x/net/context"
	"time"
)

var _ = function.NewStringFunctionClient(nil)

type grpcDispatcher struct {
	client function.StringFunctionClient
}

func (this grpcDispatcher) Dispatch(in interface{}) (interface{}, error) {
	request := fntypes.Request{Body: in.(string)}
	reply, err := this.client.Call(context.Background(), &request)
	if err != nil {
		log.Printf("Error calling gRPC server: %v", err)
		return nil, err
	}
	return reply.GetBody(), nil
}

func NewGrpcDispatcher() dispatcher.Dispatcher {
	context, _ := context.WithTimeout(context.Background(), 60 * time.Second)
	conn, err := grpc.DialContext(context, "localhost:10382", grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}

	result := grpcDispatcher{function.NewStringFunctionClient(conn)}
	return result;
}
