package core_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/projectriff/riff/pkg/core"
	"os"
)

var _ = Describe("Image prefix", func() {

	It("is computed from the user-defined Value", func() {
		result, err := core.DetermineImagePrefix("prefix", core.DockerRegistryOption(""))

		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(Equal("prefix"))
	})

	It("is computed from DockerHub ID", func() {
		result, err := core.DetermineImagePrefix("", core.DockerRegistryOption("docker"))

		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(Equal("docker.io/docker"))
	})

	It("is computed from DockerHub ID and others not set", func() {
		result, err := core.DetermineImagePrefix("", core.GoogleContainerRegistryOption(""), core.DockerRegistryOption("docker"))

		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(Equal("docker.io/docker"))
	})

	It("is computed from the GCR token file being set and based on its included project ID", func() {
		result, err := core.DetermineImagePrefix("", core.GoogleContainerRegistryOption("fixtures/gcr-creds"))

		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(Equal("gcr.io/gcp-project-id"))
	})

	It("is computed from the GCR token file being set and others not set", func() {
		result, err := core.DetermineImagePrefix("",
			core.DockerRegistryOption(""),
			&core.RegistryOption{Value: "", ImagePrefixSupplier: func(string) (string, error) { return "", nil }},
			core.GoogleContainerRegistryOption("fixtures/gcr-creds"))

		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(Equal("gcr.io/gcp-project-id"))
	})

	It("fails if the GCR token path is invalid", func() {
		_, err := core.DetermineImagePrefix("", core.GoogleContainerRegistryOption("not-a-file"))

		Expect(err).To(BeAssignableToTypeOf(&os.PathError{}))
	})
})
