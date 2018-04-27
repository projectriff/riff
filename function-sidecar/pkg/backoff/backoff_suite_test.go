package backoff

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestBackoff(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Backoff Suite")
}
