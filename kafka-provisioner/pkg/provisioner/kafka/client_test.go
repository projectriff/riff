package client_test

import (
	"github.com/Shopify/sarama"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	client "github.com/projectriff/kafka-provisioner/pkg/provisioner/kafka"
)

var _ = Describe("Kafka Client", func() {
	var (
		broker      *sarama.MockBroker
		kafkaClient client.KafkaClient
	)

	AfterEach(func() {
		broker.Close()
		err := kafkaClient.Close()
		Expect(err).NotTo(HaveOccurred())
	})

	Describe("checking for topic existence", func() {
		BeforeEach(func() {
			broker = sarama.NewMockBroker(GinkgoT(), int32(1))
			broker.SetHandlerByMap(map[string]sarama.MockResponse{
				"MetadataRequest": sarama.NewMockMetadataResponse(GinkgoT()).
					SetController(broker.BrokerID()).
					SetBroker(broker.Addr(), broker.BrokerID()).
					SetLeader("some-topic", 0, broker.BrokerID()),
			})
			kafkaClient = newKafkaClient(broker)
		})

		It("confirms when the topic has been created", func() {
			topicExists, kafkaError := kafkaClient.TopicExists("some-topic")

			Expect(kafkaError).To(BeNil())
			Expect(topicExists).To(BeTrue(), "Expected topic to exist")
		})
	})

	Describe("creating topic", func() {
		BeforeEach(func() {
			broker = sarama.NewMockBroker(GinkgoT(), int32(1))
			broker.SetHandlerByMap(map[string]sarama.MockResponse{
				"MetadataRequest": sarama.NewMockMetadataResponse(GinkgoT()).
					SetController(broker.BrokerID()).
					SetBroker(broker.Addr(), broker.BrokerID()),
				"CreateTopicsRequest": sarama.NewMockCreateTopicsResponse(GinkgoT()),
			})
			kafkaClient = newKafkaClient(broker)
		})

		It("succeeds when the topic has not been created before", func() {
			err := kafkaClient.CreateTopic("some-topic")

			Expect(err).NotTo(HaveOccurred())
		})
	})

})

func newKafkaClient(broker *sarama.MockBroker) client.KafkaClient {
	kClient, err := client.NewKafkaClient(broker.Addr())
	Expect(err).NotTo(HaveOccurred())
	return kClient
}
