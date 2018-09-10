package core_test

import (
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/projectriff/riff/pkg/core"
	"io/ioutil"
	"os"
	"path/filepath"
)

var _ = Describe("RelocateImages", func() {

	var (
		client  core.Client
		options core.RelocateImagesOptions
		err     error
	)

	BeforeEach(func() {
		client = core.NewClient(nil, nil, nil, nil)
		options.Registry = "reg"
		options.RegistryUser = "user"
	})

	JustBeforeEach(func() {
		err = client.RelocateImages(options)
	})

	Describe("manifest relocation", func() {
		AssertSuccess := func() {
			It("should write the relocated manifest to the output directory", func() {
				Expect(err).NotTo(HaveOccurred())

				// manifest should be unchanged since it contains just a filename
				actualManifest := readManifest(filepath.Join(options.Output, "manifest.yaml"))
				expectedManifest := readManifest("./fixtures/image_relocation/manifest.yaml")
				Expect(actualManifest).To(Equal(expectedManifest))

				actualYAML := readFile(filepath.Join(options.Output, "istio.yaml"))
				expectedYAML := readFile("./fixtures/image_relocation/istio_relocated.yaml")
				Expect(actualYAML).To(Equal(expectedYAML))

				actualYAML = readFile(filepath.Join(options.Output, "release.yaml"))
				expectedYAML = readFile("./fixtures/image_relocation/release_relocated.yaml")
				Expect(actualYAML).To(Equal(expectedYAML))

				actualYAML = readFile(filepath.Join(options.Output, "build.yaml"))
				expectedYAML = readFile("./fixtures/image_relocation/build_relocated.yaml")
				Expect(actualYAML).To(Equal(expectedYAML))
			})
		}

		BeforeEach(func() {
			options.Manifest = "./fixtures/image_relocation/manifest.yaml"
			options.Images = "./fixtures/image_relocation/image_manifest.yaml"
		})

		Context("when the output directory already exists", func() {
			BeforeEach(func() {
				dir, err := ioutil.TempDir("", "image-relocation-test")
				Expect(err).NotTo(HaveOccurred())
				options.Output = dir
			})

			AssertSuccess()
		})

		Context("when the output directory does not exist", func() {
			BeforeEach(func() {
				dir, err := ioutil.TempDir("", "image-relocation-test")
				Expect(err).NotTo(HaveOccurred())
				options.Output = filepath.Join(dir, "new")
			})

			AssertSuccess()
		})

		Context("when the output directory is actually a file", func() {
			BeforeEach(func() {
				dir, err := ioutil.TempDir("", "image-relocation-test")
				Expect(err).NotTo(HaveOccurred())
				options.Output = filepath.Join(dir, "file")
				err = ioutil.WriteFile(options.Output, []byte{0}, 0644)
				Expect(err).NotTo(HaveOccurred())
			})

			It("should return an appropriate error", func() {
			    Expect(err).To(MatchError(HavePrefix("output directory is a file: ")))
			})
		})
	})

	Describe("YAML file relocation", func() {
		Context("when the YAML file is specified using a file path", func() {
			BeforeEach(func() {
				options.SingleFile = "./fixtures/image_relocation/release.yaml"
				options.Images = "./fixtures/image_relocation/image_manifest.yaml"
			})

			Context("when the output path is a directory", func() {
				BeforeEach(func() {
					dir, err := ioutil.TempDir("", "image-relocation-test")
					Expect(err).NotTo(HaveOccurred())
					options.Output = dir
				})

				It("should write the relocated YAML file to the output directory", func() {
					Expect(err).NotTo(HaveOccurred())

					actual := readFile(filepath.Join(options.Output, "release.yaml"))
					expected := readFile("./fixtures/image_relocation/release_relocated.yaml")
					Expect(actual).To(Equal(expected))
				})
			})

			Context("when the output path is a file", func() {
				BeforeEach(func() {
					dir, err := ioutil.TempDir("", "image-relocation-test")
					Expect(err).NotTo(HaveOccurred())
					options.Output = filepath.Join(dir, "output.yaml")
				})

				It("should write the relocated YAML file to the output path", func() {
					Expect(err).NotTo(HaveOccurred())

					actual := readFile(options.Output)
					expected := readFile("./fixtures/image_relocation/release_relocated.yaml")
					Expect(actual).To(Equal(expected))
				})
			})
		})

		Context("when the YAML file is specified using a URL", func() {
			BeforeEach(func() {
				cwd, err := os.Getwd()
				Expect(err).NotTo(HaveOccurred())
				options.SingleFile = fmt.Sprintf("file://%s/fixtures/image_relocation/release.yaml", cwd) // local URL so test can run without network
				options.Images = "./fixtures/image_relocation/image_manifest.yaml"

				dir, err := ioutil.TempDir("", "image-relocation-test")
				Expect(err).NotTo(HaveOccurred())
				options.Output = dir
			})

			It("should write the relocated YAML file to the output directory", func() {
				Expect(err).NotTo(HaveOccurred())

				actual := readFile(filepath.Join(options.Output, "release.yaml"))
				expected := readFile("./fixtures/image_relocation/release_relocated.yaml")
				Expect(actual).To(Equal(expected))
			})
		})
	})

})

func readManifest(manifestPath string) *core.Manifest {
	manifest, err := core.NewManifest(manifestPath)
	Expect(err).NotTo(HaveOccurred())
	return manifest
}

func readFile(path string) string {
	content, err := ioutil.ReadFile(path)
	Expect(err).NotTo(HaveOccurred())
	return string(content)
}
