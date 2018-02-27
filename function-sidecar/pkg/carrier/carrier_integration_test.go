/*
 * Copyright 2017-Present the original author or authors.
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

package carrier_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"net/http"
	"os"
	"bufio"
	"github.com/bsm/sarama-cluster"
	"math/rand"
	"time"
	"github.com/Shopify/sarama"
	"github.com/projectriff/message-transport/pkg/message"
	"github.com/projectriff/message-transport/pkg/transport/kafka"
	"github.com/projectriff/function-sidecar/pkg/carrier"
	dispatch "github.com/projectriff/function-sidecar/pkg/dispatcher"
	dispatchhttp "github.com/projectriff/function-sidecar/pkg/dispatcher/http"
	"github.com/projectriff/message-transport/pkg/transport"
)

const sourceMsg = `World`
const expectedReply = `Hello World`

const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

var _ = Describe("Carrier Integration Test", func() {
	BeforeEach(func() {
		rand.Seed(time.Now().UnixNano())
	})

	It("should correctly relay messages", func(done Done) {
		broker := os.Getenv("KAFKA_BROKER")
		if broker == "" {
			Fail("Required environment variable KAFKA_BROKER was not provided")
		}

		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			bodyScanner := bufio.NewScanner(r.Body)
			if ! bodyScanner.Scan() {
				Fail("Scan of message body failed")
			}
			w.Write([]byte("Hello " + bodyScanner.Text()))
		})

		go func() {
			http.ListenAndServe(":8080", nil)
		}()

		input := randString(10)
		output := randString(10)
		group := randString(10)

		startFunctionSidecar(broker, input, output, group)

		kafkaProducer, kafkaProducerErr := kafka.NewProducer([]string{broker})
		Expect(kafkaProducerErr).NotTo(HaveOccurred())

		err := kafkaProducer.Send(input, message.NewMessage([]byte(sourceMsg), nil))
		Expect(err).NotTo(HaveOccurred())

		producerCloseErr := kafkaProducer.Close()
		Expect(producerCloseErr).NotTo(HaveOccurred())

		consumerConfig := cluster.NewConfig()
		consumerConfig.Consumer.Offsets.Initial = sarama.OffsetOldest
		group2 := randString(10)
		consumer, err := kafka.NewConsumer([]string{broker}, group2, []string{output}, consumerConfig)
		Expect(err).NotTo(HaveOccurred())
		defer consumer.Close()

		select {
		case msg, ok := <-consumer.Messages():
			if ok {
				reply := string(msg.Payload())
				Expect(reply).To(Equal(expectedReply))
				close(done)
			}
		case <-time.After(time.Second * 100):
			Fail("Timed out waiting for reply")
		}
	}, 30)
})

func randString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func startFunctionSidecar(broker string, input string, output string, group string) {
	brokers := []string{broker}

	var producer transport.Producer
	var err error
	producer, err = kafka.NewProducer(brokers)
	Expect(err).NotTo(HaveOccurred())

	consumerConfig := makeConsumerConfig()
	consumerConfig.Consumer.Offsets.Initial = sarama.OffsetOldest
	consumer, err := kafka.NewConsumer(brokers, group, []string{input}, consumerConfig)
	Expect(err).NotTo(HaveOccurred())

	dispatcher, err := dispatch.NewWrapper(dispatchhttp.NewHttpDispatcher())
	Expect(err).NotTo(HaveOccurred())

	go carrier.Run(consumer, producer, dispatcher, output)
}

func makeConsumerConfig() *cluster.Config {
	consumerConfig := cluster.NewConfig()
	consumerConfig.Consumer.Return.Errors = true
	consumerConfig.Group.Return.Notifications = true
	return consumerConfig
}
