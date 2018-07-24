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
	"github.com/bsm/sarama-cluster"
	"github.com/projectriff/riff/message-transport/pkg/transport/metrics"
)

// NewMetricsEmittingConsumer creates a producer which will delegate to a Kafka consumer and emit metrics tagged with
// the given pod id. The metrics are emitted to a metrics (Kakfa) topic which can left to default to the Riff-defined
// metrics topic. Alternatively, at most one metrics topic may be specified to use a different value than the default.
func NewMetricsEmittingConsumer(addrs []string, consumerGroup string, pod string, topics []string, consumerConfig *cluster.Config, metricsTopic ... string) (transport.Consumer, error) {
	delegate, err := kafka.NewConsumer(addrs, consumerGroup, topics, consumerConfig)
	if err != nil {
		return nil, err
	}
	metricsProducer, err := kafka.NewProducer(addrs)
	if err != nil {
		return nil, err
	}
	return metrics.NewConsumer(delegate, consumerGroup, pod, getMetricsTopic(metricsTopic...), metricsProducer), nil // TODO: pass in pod
}
