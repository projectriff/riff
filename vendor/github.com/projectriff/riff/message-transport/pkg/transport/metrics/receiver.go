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
	"github.com/projectriff/riff/message-transport/pkg/transport"
	"github.com/projectriff/riff/message-transport/pkg/message"
	"log"
	"encoding/json"
	"io"
)

func NewReceiver(consumer transport.Consumer) (*metricsReceiver) {
	producerMetricsChan := make(chan ProducerAggregateMetric)
	consumerMetricsChan := make(chan ConsumerAggregateMetric)

	go func() {
		for {
			msg, _, err := consumer.Receive()
			if err != nil {
				close(consumerMetricsChan)
				close(producerMetricsChan)
				break
			}
			pm, cm := unmarshallMetricMessage(msg)
			if pm != nil {
				producerMetricsChan <- *pm
			}
			if cm != nil {
				consumerMetricsChan <- *cm
			}
		}
	}()

	return &metricsReceiver{
		producerMetricsChan: producerMetricsChan,
		consumerMetricsChan: consumerMetricsChan,
		consumer: consumer,
	}
}

type metricsReceiver struct {
	producerMetricsChan chan ProducerAggregateMetric
	consumerMetricsChan chan ConsumerAggregateMetric
	consumer transport.Consumer
}

func (mr *metricsReceiver) ProducerMetrics() <-chan ProducerAggregateMetric {
	return mr.producerMetricsChan
}

func (mr *metricsReceiver) ConsumerMetrics() <-chan ConsumerAggregateMetric {
	return mr.consumerMetricsChan
}

const sourceHeaderKey = "source"

func unmarshallMetricMessage(msg message.Message) (*ProducerAggregateMetric, *ConsumerAggregateMetric) {
	var (
		producerMetric *ProducerAggregateMetric = nil
		consumerMetric *ConsumerAggregateMetric = nil
	)

	sourceHeader, ok := msg.Headers()[sourceHeaderKey]
	if !ok {
		log.Printf("Missing source header in metric message: %#v", msg)
	} else if len(sourceHeader) != 1 {
		log.Printf("Source header has more than one value in metric message: %#v", msg)
	} else {
		source := sourceHeader[0]

		switch source {
		case producerSource:
			var pm ProducerAggregateMetric
			err := json.Unmarshal(msg.Payload(), &pm)
			if err != nil {
				log.Printf("Error unmarshalling producer metric: %v", err)
			}
			producerMetric = &pm

		case consumerSource:
			var cm ConsumerAggregateMetric
			err := json.Unmarshal(msg.Payload(), &cm)
			if err != nil {
				log.Printf("Error unmarshalling consumer metric: %v", err)
			}
			consumerMetric = &cm

		default:
			log.Printf("Invalid source header value in metric message: %#v", msg)
		}
	}
	return producerMetric, consumerMetric
}

func (mr *metricsReceiver) Close() error {
	if consumer, ok := mr.consumer.(io.Closer); ok {
		return consumer.Close()
	}
	return nil
}