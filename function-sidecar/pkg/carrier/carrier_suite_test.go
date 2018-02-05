package carrier_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestCarrier(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Carrier Suite")
}
