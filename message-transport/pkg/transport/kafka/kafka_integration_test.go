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

package kafka_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/projectriff/message-transport/pkg/transport/kafka"
	"strings"
	"os"
	"github.com/projectriff/message-transport/pkg/transport"
	"github.com/bsm/sarama-cluster"
	"github.com/projectriff/message-transport/pkg/message"
	"github.com/Shopify/sarama"
)

var _ = Describe("Kafka Integration", func() {

	const topic = "test-topic"

	var (
		producer    transport.Producer
		consumer    transport.Consumer
		testMessage message.Message
	)

	BeforeEach(func() {
		testMessage = message.NewMessage([]byte("hello"), message.Headers{"Content-Type": []string{"bag/plastic"}})

		brokers := brokers()
		Expect(brokers).NotTo(BeEmpty())

		var err error
		producer, err = kafka.NewProducer(brokers)
		Expect(err).NotTo(HaveOccurred())

		config := cluster.NewConfig()
		config.Consumer.Offsets.Initial = sarama.OffsetOldest
		consumer, err = kafka.NewConsumer(brokers, "message-transport-integration-test", []string{topic}, config)
		Expect(err).NotTo(HaveOccurred())
	})

	It("should be able to send a message to a topic and receive it back", func() {
		err := producer.Send(topic, testMessage)
		Expect(err).NotTo(HaveOccurred())

		messages := consumer.Messages()
		receivedMessage := <-messages

		Expect(receivedMessage).To(Equal(testMessage))
	})

})

func brokers() []string {
	return strings.Split(os.Getenv("KAFKA_BROKERS"), ",")
}
