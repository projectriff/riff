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
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/projectriff/riff/function-sidecar/pkg/dispatcher"
	"time"
	"math/rand"
	"github.com/projectriff/riff/function-sidecar/pkg/dispatcher/grpc"
	"io"
	"strconv"
	"github.com/projectriff/riff/function-sidecar/pkg/dispatcher/grpc/function"
	grpc2 "google.golang.org/grpc"
	"net"
	"fmt"
	"github.com/projectriff/riff/message-transport/pkg/message"
	"github.com/projectriff/riff/kubernetes-crds/pkg/apis/projectriff.io/v1alpha1"
	"os"
	"encoding/json"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

var _ = Describe("gRPC Test", func() {

	var (
		d        dispatcher.Dispatcher
		port     int
		server   *grpc2.Server
		strategy v1alpha1.Windowing
	)

	JustBeforeEach(func() {
		port = 1024 + rand.Intn(65536-1024)
		var err error

		server = grpc2.NewServer()
		function.RegisterMessageFunctionServer(server, &myfunction{})

		l, err := net.Listen("tcp", fmt.Sprintf(":%v", port))
		Expect(err).NotTo(HaveOccurred())
		go server.Serve(l)

		s, err := json.Marshal(strategy)
		Expect(err).NotTo(HaveOccurred())
		os.Setenv("WINDOWING_STRATEGY", string(s))
		d, err = grpc.NewGrpcDispatcher(port, 100*time.Millisecond)
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		server.Stop()
	})

	Context("with unbounded (default) strategy", func() {
		It("should accumulate state", func() {
			go func() {
				GinkgoRecover()
				d.Input() <- message.NewMessage([]byte("1"), nil)
				d.Input() <- message.NewMessage([]byte("2"), nil)
				d.Input() <- message.NewMessage([]byte("3"), nil)
			}()

			m := <-d.Output()
			Expect(m.Payload()).To(Equal([]byte("1")))
			m = <-d.Output()
			Expect(m.Payload()).To(Equal([]byte("3")))
			m = <-d.Output()
			Expect(m.Payload()).To(Equal([]byte("6")))
		})
	})
	Context("with uncorrelated messages", func() {
		It("should use different streams", func() {
			go func() {
				GinkgoRecover()
				d.Input() <- message.NewMessage([]byte("1"), map[string][]string{"correlationId": {"a"}})
				d.Input() <- message.NewMessage([]byte("2"), map[string][]string{"correlationId": {"a"}})
				d.Input() <- message.NewMessage([]byte("5"), map[string][]string{"correlationId": {"b"}})
			}()

			received := make(map[string]struct{})
			m := <-d.Output()
			received[string(m.Payload())] = struct{}{}
			m = <-d.Output()
			received[string(m.Payload())] = struct{}{}
			m = <-d.Output()
			received[string(m.Payload())] = struct{}{}

			expected := map[string]struct{}{"1": {}, "3": {}, "5": {}}

			Expect(received).To(Equal(expected))
		})
	})
	Context("with size based strategy messages", func() {
		BeforeEach(func() {
			strategy = v1alpha1.Windowing{Size: 2}
		})
		It("should use different streams", func() {
			go func() {
				GinkgoRecover()
				d.Input() <- message.NewMessage([]byte("1"), nil)
				d.Input() <- message.NewMessage([]byte("2"), nil)
				d.Input() <- message.NewMessage([]byte("5"), nil)
				d.Input() <- message.NewMessage([]byte("4"), nil)
			}()

			received := make(map[string]struct{})
			m := <-d.Output()
			received[string(m.Payload())] = struct{}{}
			m = <-d.Output()
			received[string(m.Payload())] = struct{}{}
			m = <-d.Output()
			received[string(m.Payload())] = struct{}{}
			m = <-d.Output()
			received[string(m.Payload())] = struct{}{}

			expected := map[string]struct{}{"1": {}, "3": {}, "5": {}, "9": {}}

			Expect(received).To(Equal(expected))
		})
	})

	Context("with time based strategy messages", func() {
		BeforeEach(func() {
			strategy = v1alpha1.Windowing{Time: "10ms"}
		})
		It("should use different streams", func() {
			go func() {
				GinkgoRecover()
				d.Input() <- message.NewMessage([]byte("1"), nil)
				d.Input() <- message.NewMessage([]byte("2"), nil)
				time.Sleep(15 * time.Millisecond)
				d.Input() <- message.NewMessage([]byte("5"), nil)
				d.Input() <- message.NewMessage([]byte("4"), nil)
			}()

			received := make(map[string]struct{})
			m := <-d.Output()
			received[string(m.Payload())] = struct{}{}
			m = <-d.Output()
			received[string(m.Payload())] = struct{}{}
			m = <-d.Output()
			received[string(m.Payload())] = struct{}{}
			m = <-d.Output()
			received[string(m.Payload())] = struct{}{}

			expected := map[string]struct{}{"1": {}, "3": {}, "5": {}, "9": {}}

			Expect(received).To(Equal(expected))
		})
	})
	Context("with session based strategy messages", func() {
		BeforeEach(func() {
			strategy = v1alpha1.Windowing{Session: "20ms"}
		})
		It("should use different streams", func() {
			go func() {
				GinkgoRecover()
				d.Input() <- message.NewMessage([]byte("100"), nil)
				time.Sleep(5 * time.Millisecond)
				d.Input() <- message.NewMessage([]byte("50"), nil)
				time.Sleep(5 * time.Millisecond)
				d.Input() <- message.NewMessage([]byte("20"), nil)
				time.Sleep(30 * time.Millisecond)
				d.Input() <- message.NewMessage([]byte("54"), nil)
				time.Sleep(5 * time.Millisecond)
				d.Input() <- message.NewMessage([]byte("44"), nil)
			}()

			received := make(map[string]struct{})
			m := <-d.Output()
			received[string(m.Payload())] = struct{}{}
			m = <-d.Output()
			received[string(m.Payload())] = struct{}{}
			m = <-d.Output()
			received[string(m.Payload())] = struct{}{}
			m = <-d.Output()
			received[string(m.Payload())] = struct{}{}
			m = <-d.Output()
			received[string(m.Payload())] = struct{}{}

			expected := map[string]struct{}{"100": {}, "150": {}, "170": {}, "54": {}, "98": {}}

			Expect(received).To(Equal(expected))
		})
	})

})

type myfunction struct {
}

func (*myfunction) Call(stream function.MessageFunction_CallServer) error {
	sum := 0
	for {
		in, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		value, err := strconv.Atoi(string(in.Payload))
		if err != nil {
			return err
		}
		sum += value
		out := function.Message{Payload: []byte(strconv.Itoa(sum))}
		err = stream.Send(&out)
		if err != nil {
			return err
		}
	}
}
