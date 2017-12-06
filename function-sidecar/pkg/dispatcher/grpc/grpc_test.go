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

package grpc_test

import (
	"fmt"
	"io"
	"math/rand"
	"net"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/projectriff/function-sidecar/pkg/dispatcher"
	gdispatcher "github.com/projectriff/function-sidecar/pkg/dispatcher/grpc"
	"github.com/projectriff/function-sidecar/pkg/dispatcher/grpc/fntypes"
	"github.com/projectriff/function-sidecar/pkg/dispatcher/grpc/function"
	"google.golang.org/grpc"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

type myserver struct {
}

// Example handler that expects input in the form <N>:<word> and emits "Hello <word>" N times
func (*myserver) Call(stream function.StringFunction_CallServer) error {
	for {
		in, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		parts := strings.SplitN(in.Body, ":", 2)
		count, err := strconv.Atoi(parts[0])
		if err != nil {
			return err
		}
		for i := 0; i < count; i++ {
			stream.Send(&fntypes.Reply{Type: &fntypes.Reply_Body{Body: "Hello " + parts[1]}})
		}
	}
}

func Test(t *testing.T) {

	port := 1024 + rand.Intn(10000)

	server := grpc.NewServer()
	function.RegisterStringFunctionServer(server, &myserver{})

	l, err := net.Listen("tcp", fmt.Sprintf(":%v", port))
	if err != nil {
		t.Fatalf("failed to listen: %v", err)
	}
	go server.Serve(l)
	defer server.Stop()

	d, err := gdispatcher.NewGrpcDispatcher(port)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		close(d.Input())
	}()

	d.Input() <- dispatcher.Message{Payload: []byte("3:world")}

	for i := 0; i < 3; i++ {
		select {
		case ret := <-d.Output():
			if string(ret.Payload.([]byte)) != "Hello world" {
				t.Fatalf("Got %v", ret)
			}
		case <-time.After(time.Second * 5):
			t.Fatal("Timed out waiting for reply")
		}
		if i == 1 { // Send another input in the middle of processing first reply
			d.Input() <- dispatcher.Message{Payload: []byte("2:gRPC")}
		}
	}
	for i := 0; i < 2; i++ {
		select {
		case ret := <-d.Output():
			if string(ret.Payload.([]byte)) != "Hello gRPC" {
				t.Fatalf("Got %v", ret)
			}
		case <-time.After(time.Second * 5):
			t.Fatal("Timed out waiting for reply")
		}
	}

}
