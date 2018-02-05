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

package carrier_test

import (
	"github.com/projectriff/function-sidecar/pkg/carrier"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/projectriff/message-transport/pkg/transport/mocktransport"
	"github.com/projectriff/function-sidecar/pkg/dispatcher/mockdispatcher"
	"github.com/stretchr/testify/mock"
	"sync/atomic"
	"errors"
	"github.com/projectriff/message-transport/pkg/message"
)

var _ = Describe("Carrier", func() {

	const testTopic = "testtopic"

	var (
		mockConsumer     *mocktransport.Consumer
		mockProducer     *mocktransport.Producer
		numProducerSends uint32
		mockDispatcher   *mockdispatcher.Dispatcher
		consumerMessages chan message.Message
		dispatcherInput  chan message.Message
		dispatcherOutput chan message.Message

		testMessage message.Message
		replyTopic  string

		testError error
	)

	BeforeEach(func() {
		mockConsumer = &mocktransport.Consumer{}
		consumerMessages = make(chan message.Message, 1)
		mockConsumer.On("Messages").Return(receiveChan((consumerMessages)))

		mockProducer = &mocktransport.Producer{}
		numProducerSends = 0

		mockDispatcher = &mockdispatcher.Dispatcher{}
		dispatcherInput = make(chan message.Message)
		mockDispatcher.On("Input").Return(sendChan(dispatcherInput))
		dispatcherOutput = make(chan message.Message, 1)
		mockDispatcher.On("Output").Return(receiveChan(dispatcherOutput))

		testMessage = message.NewMessage([]byte("test message"), message.Headers{"h": []string{"v"}})

		testError = errors.New("To err is human")
	})

	JustBeforeEach(func() {
		carrier.Run(mockConsumer, mockProducer, mockDispatcher, replyTopic)
	})

	It("should pass a consumed message to the dispatcher", func() {
		consumerMessages <- testMessage
		Expect(<-dispatcherInput).To(Equal(testMessage))
	})

	Context("when a reply topic is provided", func() {
		BeforeEach(func() {
			replyTopic = testTopic
		})

		Context("when Send succeeds", func() {
			BeforeEach(func() {
				mockProducer.On("Send", replyTopic, testMessage).Return(nil).Run(func(mock.Arguments) {
					atomic.AddUint32(&numProducerSends, 1)
				})

				It("should pass a result from the dispatcher to the provider using the reply topic", func() {
					dispatcherOutput <- testMessage

					Eventually(func() uint32 {
						return atomic.LoadUint32(&numProducerSends)
					}).Should(Equal(uint32(1)))
				})
			})
		})

		Context("when Send return an error", func() {
			var testMessage2 message.Message

			BeforeEach(func() {
				mockProducer.On("Send", replyTopic, testMessage).Return(testError).Run(func(mock.Arguments) {
					atomic.AddUint32(&numProducerSends, 1)
				})

				testMessage2 = message.NewMessage([]byte("test message2"), message.Headers{"i": []string{"w"}})
				mockProducer.On("Send", replyTopic, testMessage2).Return(nil).Run(func(mock.Arguments) {
					atomic.AddUint32(&numProducerSends, 1)
				})
			})

			It("should continue to pass results from the dispatcher to the provider", func() {
				dispatcherOutput <- testMessage

				dispatcherOutput <- testMessage2

				Eventually(func() uint32 {
					return atomic.LoadUint32(&numProducerSends)
				}).Should(Equal(uint32(2)))
			})
		})
	})

	Context("when a reply topic is omitted", func() {
		BeforeEach(func() {
			replyTopic = ""
			mockProducer.On("Send", replyTopic, testMessage).Return(nil).Run(func(mock.Arguments) {
				atomic.AddUint32(&numProducerSends, 1)
			})
		})

		It("should not pass a result from the dispatcher to the provider", func() {
			dispatcherOutput <- testMessage

			Consistently(func() uint32 {
				return atomic.LoadUint32(&numProducerSends)
			}).Should(Equal(uint32(0)))
		})
	})
})

func receiveChan(c chan message.Message) <-chan message.Message {
	var r <-chan message.Message = c
	return r
}

func sendChan(c chan message.Message) chan<- message.Message {
	var s chan<- message.Message = c
	return s
}
