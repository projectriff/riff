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

package kafka_over_kafka

import (
	"github.com/projectriff/riff/message-transport/pkg/transport"
	"github.com/projectriff/riff/message-transport/pkg/transport/kafka"
	"github.com/projectriff/riff/message-transport/pkg/transport/metrics"
)

const DefaultMetricsTopic = "io.projectriff.message-transport.metrics"

// NewMetricsEmittingProducer creates a producer which will delegate to a Kafka producer and emit metrics tagged with
// the given producer id. The metrics are emitted to a metrics (Kafka) topic which can left to default to the Riff-defined
// metrics topic. Alternatively, at most one metrics topic may be specified to use a different value than the default.
func NewMetricsEmittingProducer(brokerAddrs []string, producerId string, metricsTopic ... string) (transport.Producer, error) {
	delegate, err := kafka.NewProducer(brokerAddrs)
	if err != nil {
		return nil, err
	}
	metricsProducer, err := kafka.NewProducer(brokerAddrs)
	if err != nil {
		return nil, err
	}
	return metrics.NewProducer(delegate, producerId, getMetricsTopic(metricsTopic...), metricsProducer), nil
}

func getMetricsTopic(metricsTopic ...string) string {
	var mTopic string
	switch len(metricsTopic) {
	case 0:
		mTopic = DefaultMetricsTopic
	case 1:
		mTopic = metricsTopic[0]
	default:
		panic("At most one metrics topic may be specified")
	}
	return mTopic
}
