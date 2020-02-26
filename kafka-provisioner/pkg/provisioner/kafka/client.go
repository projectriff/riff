package client

import (
	"github.com/Shopify/sarama"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . KafkaClient
type KafkaClient interface {
	TopicExists(topicName string) (bool, *KafkaError)
	CreateTopic(topicName string) error
	Close() error
}

type kafkaClient struct {
	Admin sarama.ClusterAdmin
}

func NewKafkaClient(brokerAddress string) (KafkaClient, error) {
	config := sarama.NewConfig()
	config.Version = sarama.V0_11_0_0
	config.ClientID = "kafka-provisioner"
	admin, err := sarama.NewClusterAdmin([]string{brokerAddress}, config)
	if err != nil {
		return nil, err
	}
	return &kafkaClient{
		Admin: admin,
	}, nil
}

type KafkaError struct {
	GeneralError error
	KError       sarama.KError
}

func (kfc *kafkaClient) TopicExists(topicName string) (bool, *KafkaError) {
	metadata, err := kfc.Admin.DescribeTopics([]string{topicName})
	if err != nil {
		return false, &KafkaError{GeneralError: err}
	}

	topicError := metadata[0].Err
	if topicError == sarama.ErrNoError {
		return true, nil
	}
	if topicError == sarama.ErrUnknownTopicOrPartition {
		return false, nil
	}
	return false, &KafkaError{KError: topicError}
}

func (kfc *kafkaClient) CreateTopic(topicName string) error {
	topicDetail := sarama.TopicDetail{NumPartitions: 1, ReplicationFactor: 1}
	return kfc.Admin.CreateTopic(topicName, &topicDetail, false)
}

func (kfc *kafkaClient) Close() error {
	return kfc.Admin.Close()
}
