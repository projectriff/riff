/*
 * Copyright 2018 the original author or authors.
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
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"

	"github.com/projectriff/riff/message-transport/pkg/message"
	"github.com/projectriff/riff/message-transport/pkg/transport/mocktransport"
	"github.com/stretchr/testify/mock"
	"github.com/projectriff/riff/message-transport/pkg/transport/stubtransport"
)

var _ = Describe("RequestsHandler", func() {

	const errorMessage = "doh!"

	var (
		mockProducer       *mocktransport.Producer
		stubConsumer       stubtransport.ConsumerStub
		mockResponseWriter *httptest.ResponseRecorder
		req                *http.Request
		testError          error
		producerErrors     chan error
		timeout            time.Duration
		gateway            *gateway
		done               chan struct{}
	)

	BeforeEach(func() {
		mockProducer = new(mocktransport.Producer)
		req = httptest.NewRequest("GET", "http://example.com", nil)
		req.URL.Path = "/requests/testtopic"
		mockResponseWriter = httptest.NewRecorder()
		testError = errors.New(errorMessage)
		timeout = time.Second * 60

		stubConsumer = stubtransport.NewConsumerStub()

		producerErrors = make(chan error, 0)
		var pErr <-chan error = producerErrors
		mockProducer.On("Errors").Return(pErr)

		mockProducer.On("Close").Return(nil)

		done = make(chan struct{})
	})

	JustBeforeEach(func() {
		gateway = New(8080, mockProducer, stubConsumer, timeout)
		go gateway.repliesLoop(done)
		gateway.requestsHandler(mockResponseWriter, req)
	})

	AfterEach(func() {
		done <- struct{}{}
	})

	Context("when the request URL is unexpected", func() {
		BeforeEach(func() {
			req.URL.Path = "/short"
		})

		It("should return a 404", func() {
			resp := mockResponseWriter.Result()
			Expect(resp.StatusCode).To(Equal(http.StatusNotFound))
		})
	})

	Context("when the request body cannot be read", func() {
		BeforeEach(func() {
			req.Body = &badReader{testError}
		})

		It("should reply with a suitable error response", func() {
			resp := mockResponseWriter.Result()
			body, _ := ioutil.ReadAll(resp.Body)

			Expect(resp.StatusCode).To(Equal(http.StatusInternalServerError))
			Expect(string(body)).To(HavePrefix(errorMessage))
		})
	})

	Context("when the request body can be read", func() {
		BeforeEach(func() {
			req.Body = ioutil.NopCloser(strings.NewReader("body"))
		})

		Context("when sending succeeds", func() {
			BeforeEach(func() {
				mockProducer.On("Send", mock.AnythingOfType("string"), mock.Anything).Run(func(args mock.Arguments) {
					msg, ok := args[1].(message.Message)
					Expect(ok).To(BeTrue())
					stubConsumer.Send(msg, "topic")
				}).Return(nil)
			})

			It("should send one message to the producer", func() {
				Expect(sendIndex(mockProducer)).To(BeNumerically(">", -1))
			})

			It("should send to the correct topic", func() {
				Expect(mockProducer.Calls[sendIndex(mockProducer)].Arguments[0]).To(Equal("testtopic"))
			})

			It("should send a message containing the correct body", func() {
				Expect(string(sentMessage(mockProducer).Payload())).To(Equal("body"))
			})

			It("should send a message with a correlation id header", func() {
				Expect(sentMessage(mockProducer).Headers()).To(HaveKey("correlationId"))
			})

			Context("when the request contains some headers that should be propagated and some that should not", func() {
				BeforeEach(func() {
					req.Header.Add("Content-Type", "text/plain")
					req.Header.Add("Accept", "application/json")
					req.Header.Add("Accept", "text/plain")
					req.Header.Add("Accept-Charset", "utf-8")
				})

				It("should send a message containing the correct headers", func() {
					headers := sentMessage(mockProducer).Headers()
					Expect(headers).To(HaveKeyWithValue("Content-Type", ConsistOf("text/plain")))
					Expect(headers).To(HaveKeyWithValue("Accept", ConsistOf("application/json", "text/plain")))
					Expect(headers).NotTo(HaveKey("Accept-Charset"))
				})
			})

			It("should reply with an OK response", func() {
				resp := mockResponseWriter.Result()
				Expect(resp.StatusCode).To(Equal(http.StatusOK))
			})
		})

		Context("when sending fails", func() {
			BeforeEach(func() {
				mockProducer.On("Send", mock.AnythingOfType("string"), mock.Anything).Return(testError)
			})

			It("should reply with a suitable error response", func() {
				resp := mockResponseWriter.Result()
				body, _ := ioutil.ReadAll(resp.Body)

				Expect(resp.StatusCode).To(Equal(http.StatusInternalServerError))
				Expect(string(body)).To(HavePrefix(errorMessage))
			})
		})

		Context("when sending succeeds but the reply takes too long", func() {
			BeforeEach(func() {
				timeout = time.Nanosecond
				mockProducer.On("Send", mock.AnythingOfType("string"), mock.Anything).Return(nil)
			})

			It("should time out with a 404", func() {
				resp := mockResponseWriter.Result()
				Expect(resp.StatusCode).To(Equal(http.StatusNotFound))
			})
		})

		Context("when sending succeeds but the producer reports an error before the reply comes back", func() {
			BeforeEach(func() {
				mockProducer.On("Send", mock.AnythingOfType("string"), mock.Anything).Run(func(args mock.Arguments) {
					msg, ok := args[1].(message.Message)
					Expect(ok).To(BeTrue())

					producerErrors <- testError

					stubConsumer.Send(msg, "topic")
				}).Return(nil)
			})

			It("should reply with OK", func() {
				resp := mockResponseWriter.Result()
				Expect(resp.StatusCode).To(Equal(http.StatusOK))
			})
		})

		Context("when sending succeeds but a reply comes back with the wrong correlation id", func() {
			BeforeEach(func() {
				timeout = 10 * time.Millisecond
				mockProducer.On("Send", mock.AnythingOfType("string"), mock.Anything).Run(func(args mock.Arguments) {
					headers := make(message.Headers)
					headers["correlationId"] = []string{""}
					stubConsumer.Send(message.NewMessage([]byte(""), headers), "banana")
				}).Return(nil)
			})

			It("should time out with a 404", func() {
				resp := mockResponseWriter.Result()
				Expect(resp.StatusCode).To(Equal(http.StatusNotFound))
			})
		})

		Context("when sending succeeds but the reply is an error", func() {
			BeforeEach(func() {
				mockProducer.On("Send", mock.AnythingOfType("string"), mock.Anything).Run(func(args mock.Arguments) {
					msg, ok := args[1].(message.Message)
					Expect(ok).To(BeTrue())
					headers := msg.Headers()
					headers["error"] = []string{"error-server-function-invocation"}
					stubConsumer.Send(message.NewMessage([]byte(""), headers), "banana")
				}).Return(nil)
			})

			It("should reply with a 500", func() {
				resp := mockResponseWriter.Result()
				Expect(resp.StatusCode).To(Equal(http.StatusInternalServerError))
			})
		})
	})
})

func sendIndex(mockProducer *mocktransport.Producer) int {
	index := -1
	for i := 0; i < len(mockProducer.Calls); i++ {
		if mockProducer.Calls[i].Method == "Send" {
			Expect(index).To(Equal(-1))
			index = i
		}
	}
	Expect(index).NotTo(Equal(-1))
	return index
}
