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
	"github.com/bsm/sarama-cluster"
	"log"
	"github.com/projectriff/riff/message-transport/pkg/message"
	"errors"
)

func NewConsumer(addrs []string, groupID string, topics []string, consumerConfig *cluster.Config) (*consumer, error) {
	clusterConsumer, err := cluster.NewConsumer(addrs, groupID, topics, consumerConfig)
	if err != nil {
		return &consumer{}, err
	}

	if consumerConfig.Consumer.Return.Errors {
		go consumeErrors(clusterConsumer)
	}
	if consumerConfig.Group.Return.Notifications {
		go consumeNotifications(clusterConsumer)
	}

	messages := make(chan messageFromTopic)

	go func(clusterConsumer *cluster.Consumer) {
		consumerMessages := clusterConsumer.Messages()
		for {
			kafkaMsg, ok := <-consumerMessages
			if ok {
				msg, err := fromKafka(kafkaMsg)
				if err != nil {
					log.Println("Failed to extract message ", err)
					continue
				}
				sendMessageFromTopic(messages, msg, kafkaMsg.Topic)
				clusterConsumer.MarkOffset(kafkaMsg, "") // mark message as processed
			}
		}
	}(clusterConsumer)

	return &consumer{
		clusterConsumer: clusterConsumer,
		messages:        messages,
	}, nil
}

func sendMessageFromTopic(messages chan<- messageFromTopic, messageWithHeaders message.Message, topic string) {
	messages <- messageFromTopic{
		Message: messageWithHeaders,
		topic:   topic,
	}
}

type messageFromTopic struct {
	message.Message
	topic string
}

type consumer struct {
	clusterConsumer *cluster.Consumer
	messages        <-chan messageFromTopic
}

func (c *consumer) Receive() (message.Message, string, error) {
	mt, ok := <-c.messages
	if !ok {
		return nil, "", errors.New("no message available")
	}
	return mt.Message, mt.topic, nil
}

func (c *consumer) Close() error {
	return c.clusterConsumer.Close()
}

func consumeErrors(consumer *cluster.Consumer) {
	for err := range consumer.Errors() {
		log.Printf("Error: %s\n", err.Error())
	}
}

func consumeNotifications(consumer *cluster.Consumer) {
	for ntf := range consumer.Notifications() {
		log.Printf("Rebalanced: %+v\n", ntf)
	}
}
