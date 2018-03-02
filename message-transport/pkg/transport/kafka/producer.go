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

package kafka

import (
	"github.com/Shopify/sarama"
	"log"
	"github.com/projectriff/riff/message-transport/pkg/message"
)

func NewProducer(brokerAddrs []string) (*producer, error) {
	asyncProducer, err := sarama.NewAsyncProducer(brokerAddrs, nil)
	if err != nil {
		return &producer{}, err
	}

	errors := make(chan error)
	go func(errChan <-chan *sarama.ProducerError) {
		for {
			errors <- <-errChan
		}
	}(asyncProducer.Errors())

	return &producer{
		asyncProducer: asyncProducer,
		errors: errors,
	}, nil
}

type producer struct {
	asyncProducer sarama.AsyncProducer
	errors        chan error
}

func (p *producer) Send(topic string, message message.Message) error {
	kafkaMsg, err := toKafka(message)
	if err != nil {
		return err
	}
	kafkaMsg.Topic = topic

	p.asyncProducer.Input() <- kafkaMsg

	return nil
}


func (p *producer) Errors() <-chan error {
	return p.errors
}

func (p *producer) Close() error {
	err := p.asyncProducer.Close()
	if err != nil {
		log.Fatalln(err)
	}
	return err
}

