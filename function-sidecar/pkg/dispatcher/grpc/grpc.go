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

	"io"

	"github.com/projectriff/riff/function-sidecar/pkg/dispatcher"
	"github.com/projectriff/riff/function-sidecar/pkg/dispatcher/grpc/function"
	"github.com/projectriff/riff/message-transport/pkg/message"
	"golang.org/x/net/context"
	"encoding/json"
	"os"
	"github.com/projectriff/riff/kubernetes-crds/pkg/apis/projectriff.io/v1alpha1"
	"reflect"
)

type grpcDispatcher struct {
	stream           function.MessageFunction_CallClient
	client           function.MessageFunctionClient
	input            chan message.Message
	output           chan message.Message
	closed           chan struct{}
	correlation      CorrelationStrategy
	windowingFactory WindowingStrategyFactory
	streams          map[interface{}]*window
}

func (this *grpcDispatcher) Input() chan<- message.Message {
	return this.input
}

func (this *grpcDispatcher) Output() <-chan message.Message {
	return this.output
}

func (this *grpcDispatcher) Closed() <-chan struct{} {
	return this.closed
}

// A CorrelationStrategy groups together messages that belong to the same "spatial" stream (the "temporal" aspect
// of grouping is taken care of by a WindowStrategy). Several invocations for different spatial groups may happen concurrently.
type CorrelationStrategy func(message.Message) interface{}

// A WindowingStrategyFactory is a function for creating a new instance of a (most certainly stateful) WindowingStrategy,
// given a new, first message for a stream.
type WindowingStrategyFactory func(message.Message) WindowingStrategy

//
type WindowingStrategy interface {
	// ShouldClose notifies this strategy that a new message has just been sent to the function, resulting in the optional
	// error passed as 2nd argument. This is a synchronous way of deciding whether or not a window should be closed.
	// ShouldClose should return true if this strategy decides that the current stream should end.
	ShouldClose(in message.Message, err error) bool

	// AsyncClose should return a non-nil channel that this strategy will close if it decides that the current window
	// should end asynchronously
	AsyncClose() <-chan struct{}
}

const correlationId = "correlationId"

var headersToForward []string

func init() {
	headersToForward = []string{correlationId}
}

type window struct {
	stream            function.MessageFunction_CallClient // the gRPC stream for this window
	windowingStrategy WindowingStrategy                   // the (stateful) strategy used to decide when to end this window
	forwardedHeaders  message.Headers                     // the set of headers to copy to each outgoing message for this window
	key               interface{}                         // the correlation key for this window
}

func (d *grpcDispatcher) handleIncoming() {

	cases := []reflect.SelectCase{
		{Chan: reflect.ValueOf(d.input), Dir: reflect.SelectRecv},
	}

	select2window := make(map[reflect.SelectCase]*window)

	for {
		chosen, value, recvOK := reflect.Select(cases)
		if chosen == 0 && recvOK {
			in := value.Interface().(message.Message)
			// TODO: problem: requires one incoming message to trigger a Call(), what about Supplier functions?
			key := d.correlation(in)
			w, ok := d.streams[key]
			if !ok {
				log.Printf("Opening new Call(),       key = %v\n", key)
				stream, err := d.client.Call(context.Background())
				if err == nil {
					windowingStrategy := d.windowingFactory(in)
					w = &window{
						stream:            stream,
						windowingStrategy: windowingStrategy,
						forwardedHeaders:  retainHeaders(in),
						key:               key,
					}
					closeChan := windowingStrategy.AsyncClose()
					if closeChan != nil {
						selectCase := reflect.SelectCase{Chan: reflect.ValueOf(closeChan), Dir: reflect.SelectRecv}
						cases = append(cases, selectCase)
						select2window[selectCase] = w
					}
					go d.handleOutgoing(w)
				} else {
					log.Printf("Error in open %v\n", err)
				}
				d.streams[key] = w
			} else {
				log.Printf("Re-using existing Call(), key = %v\n", key)
			}

			grpcMessage := toGRPC(in)
			err := w.stream.Send(grpcMessage)

			if w.windowingStrategy.ShouldClose(in, err) {
				log.Printf("Closing stream synchronously after msg, for key = %v\n", key)
				d.close(w)
				closeChan := w.windowingStrategy.AsyncClose()
				if closeChan != nil {
					selectCase := reflect.SelectCase{Chan: reflect.ValueOf(closeChan), Dir: reflect.SelectRecv}
					delete(select2window, selectCase)
				}
			} else {
				log.Printf("Keeping stream open       key = %v\n", key)
			}
		} else if chosen != 0 && !recvOK {
			// async notification to close
			selectCase := cases[chosen]
			cases = append(cases[:chosen], cases[chosen+1:]...)
			w := select2window[selectCase]
			log.Printf("Async closing of stream for key = %v", w.key)
			d.close(w)
			delete(select2window, selectCase)
		} else {
			panic("illegal state")
		}
	}
}

func retainHeaders(in message.Message) message.Headers {
	result := make(message.Headers)
	for _, h := range headersToForward {
		if v, ok := in.Headers()[h]; ok {
			result[h] = v
		}
	}
	return result
}

func (d *grpcDispatcher) handleOutgoing(w *window) {
	for {
		reply, err := w.stream.Recv()

		if err == io.EOF {
			break
		} else if err != nil {
			log.Printf("Error receiving message from function: %v", err)
			d.close(w)
			break
		}
		message := toDispatcher(reply)
		for h, v := range w.forwardedHeaders {
			message.Headers()[h] = v
		}
		d.output <- message
	}
}

func (d *grpcDispatcher) close(w *window) error {
	err := w.stream.CloseSend()
	if err != nil {
		log.Printf("Error in close %v\n", err)
	}
	delete(d.streams, w.key)
	return err
}

func (d *grpcDispatcher) Close() error {
	// shut down all streams that are in-flight
	var e error
	for _, w := range d.streams {
		if err := d.close(w); err != nil {
			e = err
		}
	}
	return e
}

func NewGrpcDispatcher(port int, timeout time.Duration) (dispatcher.Dispatcher, error) {
	ctx, _ := context.WithTimeout(context.Background(), timeout)
	conn, err := grpc.DialContext(ctx, fmt.Sprintf("localhost:%v", port), grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		return nil, err
	}

	result := &grpcDispatcher{
		client:           function.NewMessageFunctionClient(conn),
		input:            make(chan message.Message, 100),
		output:           make(chan message.Message, 100),
		closed:           make(chan struct{}),
		streams:          make(map[interface{}]*window),
		windowingFactory: unmarshallFactory(),
		correlation:      func(m message.Message) interface{} {
			if c, ok := m.Headers()[correlationId] ; ok {
				return c[0]
			}
			return nil
		},
	}
	go result.handleIncoming()

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

func unmarshallFactory() WindowingStrategyFactory {
	var w v1alpha1.Windowing
	json.Unmarshal([]byte(os.Getenv("WINDOWING_STRATEGY")), &w)

	strategies := 0
	v := reflect.ValueOf(w)
	for i := 0; i < v.NumField(); i++ {
		f := v.Field(i)
		if f.Interface() != reflect.Zero(f.Type()).Interface() {
			strategies++
		}
	}

	if strategies > 1 {
		panic("windowing strategies are mutually exclusive")
	}

	if w.Size != int32(0) {
		log.Printf("Will use windowing strategy of %v messages\n", w.Size)
		return sizeFactoryFactory(int(w.Size))
	} else if w.Time != "" {
		log.Printf("Will use time based windowing strategy of %v\n", w.Time)
		d, err := time.ParseDuration(w.Time)
		if err != nil {
			panic(err) // Assume it's validated at fn registration time
		}
		return ptimeFactoryFactory(d)
	} else if w.Session != "" {
		log.Printf("Will use session based windowing strategy of %v\n", w.Session)
		d, err := time.ParseDuration(w.Session)
		if err != nil {
			panic(err) // Assume it's validated at fn registration time
		}
		return ptimeSessionFactoryFactory(d)
	} else {
		log.Printf("Will use no windowing strategy (unbounded stream)\n")
		return noneFactoryFactory()
	}
}
