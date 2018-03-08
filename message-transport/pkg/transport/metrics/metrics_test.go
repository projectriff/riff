/*
 * Copyright 2018-Present the original author or authors.
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

package metrics_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/projectriff/riff/message-transport/pkg/transport/metrics"
	"github.com/projectriff/riff/message-transport/pkg/transport"
	"github.com/projectriff/riff/message-transport/pkg/transport/mocktransport"
	"github.com/projectriff/riff/message-transport/pkg/message"
	"errors"
	"github.com/stretchr/testify/mock"
	"github.com/projectriff/riff/message-transport/pkg/transport/stubtransport"
)

var _ bool = Describe("Metrics", func() {
	const (
		testTopic        = "test-topic"
		testMetricsTopic = "test-metrics-topic"
		testProducerId   = "a-producer"
	)

	var (
		mockMetricsProducer *mocktransport.Producer
		stubMetricsConsumer stubtransport.ConsumerStub
		metricsReceiver     metrics.MetricsReceiver
		testMessage         message.Message
		testError           error
		testError2          error
	)

	BeforeEach(func() {
		mockMetricsProducer = &mocktransport.Producer{}

		stubMetricsConsumer = stubtransport.NewConsumerStub()
		metricsReceiver = metrics.NewReceiver(stubMetricsConsumer)

		testMessage = message.NewMessage([]byte("test message"), message.Headers{})
		testError = errors.New("wat")
		testError2 = errors.New("wat else")
	})

	AfterEach(func() {
		mockMetricsProducer.AssertExpectations(GinkgoT())
	})

	Describe("Producer metrics", func() {
		var (
			mockDelegate *mocktransport.Producer
			producer     transport.Producer
			delegateErr  error
			err          error
		)

		BeforeEach(func() {
			mockDelegate = &mocktransport.Producer{}
			producer = metrics.NewProducer(mockDelegate, testProducerId, testMetricsTopic, mockMetricsProducer)
		})

		AfterEach(func() {
			mockDelegate.AssertExpectations(GinkgoT())
		})

		Describe("Send", func() {
			JustBeforeEach(func() {
				mockDelegate.On("Send", testTopic, testMessage).Return(delegateErr)
				err = producer.Send(testTopic, testMessage)
			})

			Context("when the delegate returns normally", func() {
				BeforeEach(func() {
					delegateErr = nil
				})

				Context("when the metrics producer returns normally", func() {
					BeforeEach(func() {
						mockMetricsProducer.On("Send", mock.AnythingOfType("string"), mock.Anything).Run(func(args mock.Arguments) {
							topic, ok := args[0].(string)
							Expect(ok).To(BeTrue())
							Expect(topic).To(Equal(testMetricsTopic))

							msg, ok := args[1].(message.Message)
							Expect(ok).To(BeTrue())

							stubMetricsConsumer.Send(msg, topic)
							pm, ok := <-metricsReceiver.ProducerMetrics()
							Expect(ok).To(BeTrue())
							Expect(pm.Topic).To(Equal(testTopic))
							Expect(pm.ProducerId).To(Equal(testProducerId))
							Expect(metricsReceiver.ConsumerMetrics()).To(BeEmpty())
						}).Return(nil)
					})

					It("should delegate, emit a metric, and return normally", func() {
						// metric emission is tested in the BeforeEach
						Expect(err).NotTo(HaveOccurred())
					})
				})

				Context("when the metrics producer returns an error", func() {
					BeforeEach(func() {
						mockMetricsProducer.On("Send", mock.AnythingOfType("string"), mock.Anything).Run(func(args mock.Arguments) {
							topic, ok := args[0].(string)
							Expect(ok).To(BeTrue())
							Expect(topic).To(Equal(testMetricsTopic))

							msg, ok := args[1].(message.Message)
							Expect(ok).To(BeTrue())

							stubMetricsConsumer.Send(msg, topic)
							pm, ok := <-metricsReceiver.ProducerMetrics()
							Expect(ok).To(BeTrue())
							Expect(pm.Topic).To(Equal(testTopic))
							Expect(pm.ProducerId).To(Equal(testProducerId))
							Expect(metricsReceiver.ConsumerMetrics()).To(BeEmpty())
						}).Return(testError)
					})

					It("should delegate, ignore the metrics producer error, and return normally", func() {
						Expect(err).NotTo(HaveOccurred())
					})
				})
			})

			Context("when the delegate returns an error", func() {
				BeforeEach(func() {
					delegateErr = testError
				})

				It("should propagate the error", func() {
					Expect(err).To(Equal(testError))
				})
			})
		})

		Describe("Errors", func() {
			var testChan <-chan error

			BeforeEach(func() {
				testChan = make(chan error)
				mockDelegate.On("Errors").Return(testChan)
			})

			It("should delegate", func() {
				errChan := producer.Errors()
				Expect(errChan).To(Equal(testChan))
			})
		})

		Describe("Close", func() {
			Context("when the delegate producer and the metrics producer close successfully", func() {
				BeforeEach(func() {
					mockDelegate.On("Close").Return(nil)
					mockMetricsProducer.On("Close").Return(nil)
				})

				It("should delegate and return successfully", func() {
					err := producer.Close()
					Expect(err).NotTo(HaveOccurred())
				})
			})

			Context("when the delegate producer returns an error from close and the metrics producer closes successfully", func() {
				BeforeEach(func() {
					mockDelegate.On("Close").Return(testError)
					mockMetricsProducer.On("Close").Return(nil)
				})

				It("should delegate and return the error", func() {
					err := producer.Close()
					Expect(err).To(Equal(testError))
				})
			})

			Context("when the delegate producer closes successfully and the metrics producer returns an error from close", func() {
				BeforeEach(func() {
					mockDelegate.On("Close").Return(nil)
					mockMetricsProducer.On("Close").Return(testError)
				})

				It("should delegate and return the error", func() {
					err := producer.Close()
					Expect(err).To(Equal(testError))
				})
			})

			Context("when both the delegate producer and the metrics producer return errors from close", func() {
				BeforeEach(func() {
					mockDelegate.On("Close").Return(testError)
					mockMetricsProducer.On("Close").Return(testError2)
				})

				It("should delegate and return the error from the delegate producer", func() {
					err := producer.Close()
					Expect(err).To(Equal(testError))
				})
			})
		})
	})

	Describe("Consumer metrics", func() {
		const (
			testFunction = "test-function"
			testPod      = "test-pod"
		)

		var (
			mockDelegate  *mocktransport.Consumer
			consumer      transport.Consumer
			receivedMsg   message.Message
			receivedTopic string
			receivedErr   error
		)

		BeforeEach(func() {
			mockDelegate = &mocktransport.Consumer{}
			consumer = metrics.NewConsumer(mockDelegate, testFunction, testPod, testMetricsTopic, mockMetricsProducer)
		})

		AfterEach(func() {
			mockDelegate.AssertExpectations(GinkgoT())
		})

		Describe("Receive", func() {
			JustBeforeEach(func() {
				receivedMsg, receivedTopic, receivedErr = consumer.Receive()
			})

			Context("when the delegate returns a message", func() {
				BeforeEach(func() {
					mockDelegate.On("Receive").Return(testMessage, testTopic, nil)
				})

				Context("when the metrics producer returns normally", func() {
					BeforeEach(func() {
						mockMetricsProducer.On("Send", mock.AnythingOfType("string"), mock.Anything).Run(func(args mock.Arguments) {
							topic, ok := args[0].(string)
							Expect(ok).To(BeTrue())
							Expect(topic).To(Equal(testMetricsTopic))

							msg, ok := args[1].(message.Message)
							Expect(ok).To(BeTrue())

							stubMetricsConsumer.Send(msg, topic)
							cm, ok := <-metricsReceiver.ConsumerMetrics()
							Expect(ok).To(BeTrue())
							Expect(cm.Topic).To(Equal(testTopic))
							Expect(cm.Function).To(Equal(testFunction))
							Expect(cm.Pod).To(Equal(testPod))
							Expect(metricsReceiver.ProducerMetrics()).To(BeEmpty())
						}).Return(nil)
					})

					It("should delegate and emit a metric", func() {
						// metric emission is tested in the BeforeEach
						Expect(receivedMsg).To(Equal(testMessage))
						Expect(receivedTopic).To(Equal(testTopic))
						Expect(receivedErr).NotTo(HaveOccurred())
					})
				})

				Context("when the metrics producer returns an error", func() {
					BeforeEach(func() {
						mockMetricsProducer.On("Send", mock.AnythingOfType("string"), mock.Anything).Run(func(args mock.Arguments) {
							topic, ok := args[0].(string)
							Expect(ok).To(BeTrue())
							Expect(topic).To(Equal(testMetricsTopic))

							msg, ok := args[1].(message.Message)
							Expect(ok).To(BeTrue())

							stubMetricsConsumer.Send(msg, topic)
							cm, ok := <-metricsReceiver.ConsumerMetrics()
							Expect(ok).To(BeTrue())
							Expect(cm.Topic).To(Equal(testTopic))
							Expect(cm.Function).To(Equal(testFunction))
							Expect(cm.Pod).To(Equal(testPod))
							Expect(metricsReceiver.ProducerMetrics()).To(BeEmpty())
						}).Return(testError)
					})

					It("should delegate and ignore the metrics producer error", func() {
						// metric emission is tested in the BeforeEach
						Expect(receivedMsg).To(Equal(testMessage))
						Expect(receivedTopic).To(Equal(testTopic))
						Expect(receivedErr).NotTo(HaveOccurred())
					})
				})
			})

			Context("when the delegate returns an error", func() {
				BeforeEach(func() {
					mockDelegate.On("Receive").Return(nil, "", testError)
				})

				It("should return the error", func() {
					// metric non-emission is tested by the absence of a mockMetricsProducer expectation
					Expect(receivedErr).To(Equal(testError))
				})
			})
		})

		Describe("Close", func() {
			Context("when the delegate consumer and the metrics producer close successfully", func() {
				BeforeEach(func() {
					mockDelegate.On("Close").Return(nil)
					mockMetricsProducer.On("Close").Return(nil)
				})

				It("should delegate and return successfully", func() {
					err := consumer.Close()
					Expect(err).NotTo(HaveOccurred())
				})
			})

			Context("when the delegate consumer returns an error from close and the metrics producer closes successfully", func() {
				BeforeEach(func() {
					mockDelegate.On("Close").Return(testError)
					mockMetricsProducer.On("Close").Return(nil)
				})

				It("should delegate and return the error", func() {
					err := consumer.Close()
					Expect(err).To(Equal(testError))
				})
			})

			Context("when the delegate consumer closes successfully and the metrics producer returns an error from close", func() {
				BeforeEach(func() {
					mockDelegate.On("Close").Return(nil)
					mockMetricsProducer.On("Close").Return(testError)
				})

				It("should delegate and return the error", func() {
					err := consumer.Close()
					Expect(err).To(Equal(testError))
				})
			})

			Context("when both the delegate consumer and the metrics producer return errors from close", func() {
				BeforeEach(func() {
					mockDelegate.On("Close").Return(testError)
					mockMetricsProducer.On("Close").Return(testError2)
				})

				It("should delegate and return the error from the delegate producer", func() {
					err := consumer.Close()
					Expect(err).To(Equal(testError))
				})
			})
		})
	})
})
