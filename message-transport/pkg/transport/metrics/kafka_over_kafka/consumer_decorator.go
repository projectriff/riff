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

func NewMetricsEmittingConsumer(addrs []string, function string, pod string, topics []string, consumerConfig *cluster.Config, metricsTopic ... string) (transport.Consumer, error) {
	// Create a Kafka consumer to delegate to using the function name to identify the consumer group.
	delegate, err := kafka.NewConsumer(addrs, function, topics, consumerConfig)
	if err != nil {
		return nil, err
	}
	metricsProducer, err := kafka.NewProducer(addrs)
	if err != nil {
		return nil, err
	}
	return metrics.NewConsumer(delegate, function, pod, getMetricsTopic(metricsTopic...), metricsProducer), nil // TODO: pass in pod
}
