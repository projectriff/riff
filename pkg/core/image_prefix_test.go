package core_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/projectriff/riff/pkg/core"
	"os"
)


var _ = Describe("Image prefix", func() {

	It("is computed from the user-defined value", func() {
		result, err := core.DetermineImagePrefix("prefix", "", "")

		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(Equal("prefix"))
	})

	It("is inferred from the DockerHub ID being set", func() {
		result, err := core.DetermineImagePrefix("", "docker", "")

		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(Equal("docker.io/docker"))
	})

	It("is inferred from the GCR token file being set and based on its included project ID", func() {
		result, err := core.DetermineImagePrefix("", "", "fixtures/gcr-creds")

		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(Equal("gcr.io/gcp-project-id"))
	})

	It("fails if the GCR token path is invalid", func() {
		_, err := core.DetermineImagePrefix("", "", "not-a-file")

		Expect(err).To(BeAssignableToTypeOf(&os.PathError{}))
	})
})