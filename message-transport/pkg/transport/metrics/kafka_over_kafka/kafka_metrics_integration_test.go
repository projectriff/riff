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

package kafka_over_kafka_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/projectriff/riff/message-transport/pkg/transport/metrics/kafka_over_kafka"
	"strings"
	"os"
	"github.com/projectriff/riff/message-transport/pkg/transport"
	"github.com/bsm/sarama-cluster"
	"github.com/projectriff/riff/message-transport/pkg/message"
	"github.com/Shopify/sarama"
	"time"
	"fmt"
	"github.com/projectriff/riff/message-transport/pkg/transport/metrics"
)

var _ = Describe("Kafka Metrics Integration", func() {
	const (
		testPod        = "test-pod"
		testProducerId = "some-producer"
	)

	var (
		testTopic         string
		producer          transport.Producer
		consumer          transport.Consumer
		testMessage       message.Message
		brokers           []string
		config            *cluster.Config
		consumingFunction string
		metricsTopic      string
		metricsGroupId    string
	)

	BeforeEach(func() {
		testTopic = fmt.Sprintf("topic-%d", time.Now().Nanosecond())

		testMessage = message.NewMessage([]byte("hello"), message.Headers{"Content-Type": []string{"bag/plastic"}})

		brokers = getBrokers()
		Expect(brokers).NotTo(BeEmpty())

		metricsTopic = fmt.Sprintf("metrics-topic-%d", time.Now().Nanosecond())

		var err error
		producer, err = kafka_over_kafka.NewMetricsEmittingProducer(brokers, testProducerId, metricsTopic)
		Expect(err).NotTo(HaveOccurred())

		config = cluster.NewConfig()

		// Use "oldest" initial offset in case there is a race between the asynchronous construction of the consumer
		// machinery and the producer writing the “new” message.
		config.Consumer.Offsets.Initial = sarama.OffsetOldest

		// Use a fresh function name, since that will determine the function's consumer group and then runs in close
		// succession won't suffer from Kafka broker delays due to consumers coming and going in the same group.
		consumingFunction = fmt.Sprintf("consuming-function-%d", time.Now().Nanosecond())

		// Use a fresh group id so that runs in close succession won't suffer from Kafka broker delays
		// due to consumers coming and going in the same group
		metricsGroupId = fmt.Sprintf("metrics-group-%d", time.Now().Nanosecond())

		consumer, err = kafka_over_kafka.NewMetricsEmittingConsumer(brokers, consumingFunction, testPod, []string{testTopic}, config, metricsTopic)
		Expect(err).NotTo(HaveOccurred())
	})

	It("should be able to send a message to a topic, receive it back, and emit the correct metrics", func() {
		err := producer.Send(testTopic, testMessage)
		Expect(err).NotTo(HaveOccurred())

		msg, topic, err := consumer.Receive()
		Expect(err).NotTo(HaveOccurred())
		Expect(msg).To(Equal(testMessage))
		Expect(topic).To(Equal(testTopic))

		metricsReceiver, err := kafka_over_kafka.NewMetricsReceiver(brokers, metricsGroupId, config, metricsTopic)
		Expect(err).NotTo(HaveOccurred())

		producerMetrics := metricsReceiver.ProducerMetrics()
		Expect(<-producerMetrics).To(Equal(metrics.ProducerAggregateMetric{Topic: testTopic, ProducerId: testProducerId, Count: 1}))
		Expect(producerMetrics).To(BeEmpty())

		consumerMetrics := metricsReceiver.ConsumerMetrics()
		Expect(<-consumerMetrics).To(Equal(metrics.ConsumerAggregateMetric{Topic: testTopic, Function: consumingFunction, Pod: testPod, Count: 1}))
		Expect(consumerMetrics).To(BeEmpty())
	})
})

func getBrokers() []string {
	return strings.Split(os.Getenv("KAFKA_BROKERS"), ",")
}
