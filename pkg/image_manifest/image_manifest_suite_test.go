package image_manifest_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestImageManifest(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "ImageManifest Suite")
}
