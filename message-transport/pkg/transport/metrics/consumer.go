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

package metrics

import (
	"github.com/projectriff/riff/message-transport/pkg/message"
	"github.com/projectriff/riff/message-transport/pkg/transport"
	"encoding/json"
	"log"
	"time"
	"io"
)

const consumerSource = "consumer"

// ConsumerAggregateMetric represents the reception of a number of messages from a topic by a consumer group in a time interval.
type ConsumerAggregateMetric struct {
	Topic         string
	ConsumerGroup string
	Pod           string
	Interval      time.Duration
	Count         int32
}

// NewConsumer decorates the given delegate to send consumer metrics for the given consumer group and pod to the given topic
// using the given metrics producer. The given pod can be any unique identifier for the pod which will use this consumer
// and can be pod instance-specific (that is, it need not carry across when a pod is restarted).
func NewConsumer(delegate transport.Consumer, consumerGroup string, pod string, metricsTopic string, metricsProducer transport.Producer) *consumer {
	return &consumer{
		consumerGroup:   consumerGroup,
		pod:             pod,
		delegate:        delegate,
		metricsTopic:    metricsTopic,
		metricsProducer: metricsProducer,
	}
}

func (c *consumer) createConsumerMetricMessage(topic string) message.Message {
	metric, err := json.Marshal(ConsumerAggregateMetric{
		Topic:         topic,
		ConsumerGroup: c.consumerGroup,
		Pod:           c.pod,
		// TODO: aggregate metrics into suitable intervals
		Interval: time.Duration(0),
		Count:    1,
	})
	if err != nil { // should never happen
		panic(err)
	}
	return message.NewMessage(metric, message.Headers{sourceHeaderKey: []string{consumerSource}})
}

type consumer struct {
	consumerGroup   string
	pod             string
	delegate        transport.Consumer
	metricsTopic    string
	metricsProducer transport.Producer
}

func (c *consumer) Receive() (message.Message, string, error) {
	// TODO: emit "ready to receive" metric consumer group/consumer/timestamp
	m, t, err := c.delegate.Receive()
	if err != nil {
		return nil, "", err
	}

	metricsErr := c.metricsProducer.Send(c.metricsTopic, c.createConsumerMetricMessage(t))
	if metricsErr != nil {
		log.Printf("Failed to send consumer metrics: %v", metricsErr)
		return nil, "", metricsErr
	}

	return m, t, nil
}

func (c *consumer) Close() error {
	var err error = nil
	if delegate, ok := c.delegate.(io.Closer); ok {
		err = delegate.Close()
	}

	var err2 error = nil
	if metricsProducer, ok := c.metricsProducer.(io.Closer); ok {
		err2 = metricsProducer.Close()
	}
	if err != nil {
		return err
	}
	return err2
}
