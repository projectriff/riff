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

package stubtransport

import (
	"github.com/projectriff/riff/message-transport/pkg/message"
	"github.com/projectriff/riff/message-transport/pkg/transport"
)

// This stub is for use in tests.
type ConsumerStub interface {
	transport.Consumer
	Send(msg message.Message, topic string)
}

type messageFromTopic struct {
	message.Message
	topic string
}

type consumerStub struct {
	ch chan messageFromTopic
}

func NewConsumerStub() *consumerStub {
	return &consumerStub{
		ch: make(chan messageFromTopic),
	}
}

func (stub *consumerStub) Send(msg message.Message, topic string) {
	stub.ch <- messageFromTopic{
		Message: msg,
		topic:   topic,
	}
}

func (stub *consumerStub) Receive() (message.Message, string, error) {
	msgFromTopic := <-stub.ch
	return msgFromTopic.Message, msgFromTopic.topic, nil
}

func (stub *consumerStub) Close() error {
	return nil
}

