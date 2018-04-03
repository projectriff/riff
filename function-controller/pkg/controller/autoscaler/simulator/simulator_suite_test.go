package main

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestSimulator(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Simulator Suite")
}
