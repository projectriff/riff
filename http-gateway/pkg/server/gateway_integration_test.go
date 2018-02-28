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

package server_test

import (
	"bytes"
	"fmt"
	"io"
	"math/rand"
	"net"
	"net/http"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/projectriff/riff/http-gateway/pkg/server"
	"github.com/projectriff/riff/message-transport/pkg/message"
	"github.com/projectriff/riff/message-transport/pkg/transport/mocktransport"
	"github.com/stretchr/testify/mock"
)

var _ = Describe("HTTP Gateway", func() {
	var (
		gw               server.Gateway
		mockProducer     *mocktransport.Producer
		mockConsumer     *mocktransport.Consumer
		port             int
		timeout          time.Duration
		done             chan struct{}
		consumerMessages chan message.Message
		producerErrors   chan error
	)

	BeforeEach(func() {
		mockProducer = new(mocktransport.Producer)
		mockConsumer = new(mocktransport.Consumer)

		consumerMessages = make(chan message.Message, 1)
		var cMsg <-chan message.Message = consumerMessages
		mockConsumer.On("Messages").Return(cMsg)

		producerErrors = make(chan error, 0)
		var pErr <-chan error = producerErrors
		mockProducer.On("Errors").Return(pErr)

		timeout = 6 * time.Second
		done = make(chan struct{})
	})

	JustBeforeEach(func() {
		port = 1024 + rand.Intn(32768-1024)
		gw = server.New(port, mockProducer, mockConsumer, timeout)

		gw.Run(done)

		waitForHttpGatewayToBeReady()
	})

	AfterEach(func() {
		done <- struct{}{}
	})

	It("should request/reply OK", func() {

		mockProducer.On("Send", "foo", mock.Anything).Return(nil).Run(func(args mock.Arguments) {
			defer GinkgoRecover()
			msg := args[1].(message.Message)
			consumerMessages <- message.NewMessage([]byte("hello "+string(msg.Payload())),
				message.Headers{server.CorrelationId: msg.Headers()[server.CorrelationId],
					"Content-Type": []string{"bag/plastic"},
				})
			Eventually(msg.Headers()["Content-Type"]).Should(Equal([]string{"text/solid"}))
			Eventually(msg.Headers()["Not-Propagated-Header"]).Should(BeNil())
		})

		resp := doRequest(port, "foo", bytes.NewBufferString("world"), "Content-Type", "text/solid", "Not-Propagated-Header", "secret")

		b := make([]byte, 11)
		resp.Body.Read(b)

		Eventually(b).Should(Equal([]byte("hello world")))
		Eventually(resp.Header.Get(server.CorrelationId)).Should(BeZero())
		Eventually(resp.Header.Get("Content-Type")).Should(Equal("bag/plastic"))

		defer resp.Body.Close()
	})

	It("should accept messages and fire&forget", func() {

		mockProducer.On("Send", "bar", mock.Anything).Return(nil).Run(func(args mock.Arguments) {
			defer GinkgoRecover()
			msg := args[1].(message.Message)
			Eventually(msg.Payload()).Should(Equal([]byte("world")))
			Eventually(msg.Headers()["Content-Type"]).Should(Equal([]string{"text/solid"}))
			Eventually(msg.Headers()["Not-Propagated-Header"]).Should(BeNil())
		})

		resp := doMessage(port, "bar", bytes.NewBufferString("world"), "Content-Type", "text/solid", "Not-Propagated-Header", "secret")

		Eventually(resp.StatusCode).Should(Equal(200))

		defer resp.Body.Close()
	})
})

func doRequest(port int, topic string, body io.Reader, headerKV ...string) *http.Response {
	return post(port, "/requests/"+topic, body, headerKV...)
}

func doMessage(port int, topic string, body io.Reader, headerKV ...string) *http.Response {
	return post(port, "/messages/"+topic, body, headerKV...)
}

func post(port int, path string, body io.Reader, headerKV ...string) *http.Response {
	client := http.Client{}
	req, err := http.NewRequest("POST", fmt.Sprintf("http://localhost:%v%v", port, path), body)
	if err != nil {
		panic(err)
	}

	for i := 0; i < len(headerKV); i += 2 {
		req.Header.Add(headerKV[i], headerKV[i+1])
	}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}

	return resp
}

func waitForHttpGatewayToBeReady() {
	attempts := 20
	durationBetweenAttempts := time.Millisecond * 100

	for i := 0; i < attempts; i++ {
		_, err := net.Dial("tcp", ":http")
		if err != nil {
			fmt.Print(".")
			time.Sleep(durationBetweenAttempts)
			continue
		}

		break
	}
}
