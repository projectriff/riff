package client_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestKafkaClient(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Kafka Client Suite")
}
