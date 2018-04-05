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
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/projectriff/riff/message-transport/pkg/message"
	"github.com/projectriff/riff/message-transport/pkg/transport/mocktransport"
	"github.com/stretchr/testify/mock"
)

var _ = Describe("MessagesHandler", func() {

	const errorMessage = "doh!"

	var (
		mockProducer       *mocktransport.Producer
		mockConsumer       *mocktransport.Consumer
		mockResponseWriter *httptest.ResponseRecorder
		req                *http.Request
		testError          error
		gateway            *gateway
	)

	BeforeEach(func() {
		mockProducer = new(mocktransport.Producer)
		req = httptest.NewRequest("GET", "http://example.com", nil)
		req.URL.Path = "/messages/testtopic"
		mockResponseWriter = httptest.NewRecorder()
		testError = errors.New(errorMessage)

		gateway = New(8080, mockProducer, mockConsumer, 60*time.Second, &happyRiffTopicExistenceChecker{testName: "testtopic"})
	})

	JustBeforeEach(func() {
		gateway.messagesHandler(mockResponseWriter, req)
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

	Context("when the request refers to a non-existent Riff Topic", func() {
		BeforeEach(func() {
			req.URL.Path = "/messages/nosuchtopicexists"
		})

		It("should return a 404", func() {
			resp := mockResponseWriter.Result()
			Expect(resp.StatusCode).To(Equal(http.StatusNotFound))
		})
	})

	Context("When an unexpected error occurs while looking up a Riff topic", func() {
		BeforeEach(func() {
			gateway.riffTopicExistenceChecker = &errorRiffTopicExistenceChecker{}
		})

		AfterEach(func() {
			gateway.riffTopicExistenceChecker = &happyRiffTopicExistenceChecker{}
		})

		It("should return a 500", func() {
			resp := mockResponseWriter.Result()
			Expect(resp.StatusCode).To(Equal(http.StatusInternalServerError))
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
				mockProducer.On("Send", mock.AnythingOfType("string"), mock.Anything).Return(nil)
			})

			It("should send one message to the producer", func() {
				Expect(len(mockProducer.Calls)).To(Equal(1))
				Expect(mockProducer.Calls[0].Method).To(Equal("Send"))
			})

			It("should send to the correct topic", func() {
				Expect(mockProducer.Calls[0].Arguments[0]).To(Equal("testtopic"))
			})

			It("should send a message containing the correct body", func() {
				Expect(string(sentMessage(mockProducer).Payload())).To(Equal("body"))
			})

			It("should send a message with no headers", func() {
				Expect(sentMessage(mockProducer).Headers()).To(BeEmpty())
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

				Context("when there is an HTTP header whitelist", func() {
					AfterEach(func() {
						os.Unsetenv("RIFF_HTTP_HEADERS_WHITELIST")
					})

					Context("when the whitelist has one entry", func() {
						BeforeEach(func() {
							req.Header.Add("Only-Header-In-Whitelist", "PassedThrough")
							os.Setenv("RIFF_HTTP_HEADERS_WHITELIST", "Only-Header-In-Whitelist")
						})

						It("should allow non-default headers found in a whitelist", func() {
							headers := sentMessage(mockProducer).Headers()
							Expect(headers).To(HaveKeyWithValue("Only-Header-In-Whitelist", ConsistOf("PassedThrough")))
						})
					})

					Context("when the whitelist has multiple entries", func() {
						BeforeEach(func() {
							req.Header.Add("First-hEader-In-Whitelist", "FirstPassedThrough") // Note non canonical case here
							req.Header.Add("Second-Header-In-Whitelist", "SecondPassedThrough")
							os.Setenv("RIFF_HTTP_HEADERS_WHITELIST", "First-Header-In-whiteList,Second-Header-In-Whitelist") // and here
						})

						It("should allow non-default headers found in a whitelist", func() {
							headers := sentMessage(mockProducer).Headers()
							Expect(headers).To(HaveKeyWithValue("First-Header-In-Whitelist", ConsistOf("FirstPassedThrough")))
							Expect(headers).To(HaveKeyWithValue("Second-Header-In-Whitelist", ConsistOf("SecondPassedThrough")))
						})
					})

					Context("when the whitelist is blank", func() {
						BeforeEach(func() {
							req.Header.Add("This-Header-Wont-Get-Through", "")
							os.Setenv("RIFF_HTTP_HEADERS_WHITELIST", "")
						})

						It("only allows the default headers", func() {
							headers := sentMessage(mockProducer).Headers()
							Expect(headers).NotTo(HaveKey("This-Header-Wont-Get-Through"))
						})
					})
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
	})
})

func sentMessage(mockProducer *mocktransport.Producer) message.Message {
	msg, ok := mockProducer.Calls[sendIndex(mockProducer)].Arguments[1].(message.Message)
	Expect(ok).To(BeTrue())
	return msg
}

type badReader struct {
	readErr error
}

func (br *badReader) Read(p []byte) (n int, err error) {
	return 0, br.readErr
}

func (*badReader) Close() error {
	return nil
}

type happyRiffTopicExistenceChecker struct {
	testName string
}

func (th *happyRiffTopicExistenceChecker) TopicExists(topicName string) (bool, error) {
	return topicName == th.testName, nil
}

type errorRiffTopicExistenceChecker struct{}

func (th *errorRiffTopicExistenceChecker) TopicExists(topicName string) (bool, error) {
	return false, errors.New("test error")
}
