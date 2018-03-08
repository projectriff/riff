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
	"time"
	"io"
)

const producerSource = "producer"

// ProducerAggregateMetric represents the transmission of a number of messages to a topic by a producer in a time interval.
type ProducerAggregateMetric struct {
	Topic      string
	ProducerId string
	Interval   time.Duration
	Count      int32
}

// NewProducer decorates the given delegate to send producer metrics for the given producer id to the given topic using the
// given metrics producer. The given producer id can be any unique identifier for the producer and can be pod
// instance-specific (that is, it need not carry across when a pod is restarted).
func NewProducer(delegate transport.Producer, producerId string, metricsTopic string, metricsProducer transport.Producer) *producer {
	return &producer{
		delegate:        delegate,
		producerId:      producerId,
		metricsTopic:    metricsTopic,
		metricsProducer: metricsProducer,
	}
}

type producer struct {
	delegate        transport.Producer
	producerId      string
	metricsTopic    string
	metricsProducer transport.Producer
}

func (p *producer) Send(topic string, msg message.Message) error {
	err := p.delegate.Send(topic, msg)
	if err == nil {
		metricsErr := p.metricsProducer.Send(p.metricsTopic, p.createProducerMetricMessage(topic))
		if metricsErr != nil {
			log.Printf("Failed to send producer metrics: %v", metricsErr)
			return metricsErr
		}
	}
	return err
}

func (p *producer) createProducerMetricMessage(topic string) message.Message {
	metric, err := json.Marshal(ProducerAggregateMetric{
		Topic: topic,
		ProducerId: p.producerId,
		// TODO: aggregate metrics into suitable intervals
		Interval: time.Duration(0),
		Count: 1,
	})
	if err != nil { // should never happen
		panic(err)
	}
	return message.NewMessage(metric, message.Headers{"source": []string{producerSource}})
}

func (p *producer) Errors() <-chan error {
	return p.delegate.Errors()
}

func (p *producer) Close() error {
	var err error = nil
	if delegate, ok := p.delegate.(io.Closer); ok {
		err = delegate.Close()
	}

	var err2 error = nil
	if metricsProducer, ok := p.metricsProducer.(io.Closer); ok {
		err2 = metricsProducer.Close()
	}
	if err != nil {
		return err
	}
	return err2

}
