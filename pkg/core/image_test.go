package core

import (
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
)

var _ = Describe("RelocateImages", func() {

	var (
		client  Client
		options RelocateImagesOptions
		err     error
	)

	BeforeEach(func() {
		client = NewClient(nil, nil, nil, nil)
		options.Registry = "reg"
		options.RegistryUser = "user"
	})

	JustBeforeEach(func() {
		err = client.RelocateImages(options)
	})

	Describe("manifest relocation", func() {
		Context("when there are no collisions in input file names", func() {
			AssertSuccess := func() {
				It("should write the relocated manifest to the output directory", func() {
					Expect(err).NotTo(HaveOccurred())

					actualManifest := readManifest(filepath.Join(options.Output, "manifest.yaml"))
					Expect(len(actualManifest.Istio)).To(Equal(1))

					actualYAML := readFileOk(filePath(actualManifest.Istio[0]))
					expectedYAML := readFileOk("./fixtures/image_relocation/istio_relocated.yaml")
					Expect(actualYAML).To(Equal(expectedYAML))

					Expect(len(actualManifest.Knative)).To(Equal(1))
					actualYAML = readFileOk(filePath(actualManifest.Knative[0]))
					expectedYAML = readFileOk("./fixtures/image_relocation/release_relocated.yaml")
					Expect(actualYAML).To(Equal(expectedYAML))

					Expect(len(actualManifest.Namespace)).To(Equal(1))
					actualYAML = readFileOk(filePath(actualManifest.Namespace[0]))
					expectedYAML = readFileOk("./fixtures/image_relocation/build_relocated.yaml")
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

		Context("when there are collisions in input file names", func() {
			BeforeEach(func() {
				options.Manifest = "./fixtures/image_relocation/colliding/manifest.yaml"
				options.Images = "./fixtures/image_relocation/image_manifest.yaml"

				dir, err := ioutil.TempDir("", "image-relocation-test")
				Expect(err).NotTo(HaveOccurred())
				options.Output = dir
			})

			It("should avoid unintended collisions in the output manifest", func() {
				Expect(err).NotTo(HaveOccurred())

				actualManifest := readManifest(filepath.Join(options.Output, "manifest.yaml"))

				istios := actualManifest.Istio
				Expect(len(istios)).To(Equal(1))
				actualIstio := readURLOk( istios[0])
				expectedIstio := readFileOk("./fixtures/image_relocation/istio_relocated.yaml")
				Expect(actualIstio).To(Equal(expectedIstio))

				knatives := actualManifest.Knative
				Expect(len(knatives)).To(Equal(1))
				actualBuild1 := readURLOk(knatives[0])
				expectedBuild := readFileOk("./fixtures/image_relocation/build_relocated.yaml")
				Expect(actualBuild1).To(Equal(expectedBuild))

				namespaces := actualManifest.Namespace
				Expect(len(namespaces)).To(Equal(1))
				actualBuild2 := readURLOk(namespaces[0])
				Expect(actualBuild2).To(Equal(expectedBuild))
			})

			Context("when the flatteners fail to prevent collisions", func() {
				var oldFlatteners []uriFlattener
			    BeforeEach(func() {
			    	oldFlatteners = flatteners
			        flatteners = []uriFlattener{baseFlattener}
			    })

				AfterEach(func() {
				    flatteners = oldFlatteners
				    oldFlatteners = nil
				})

			    It("should return a suitable error", func() {
			        Expect(err).To(MatchError("cannot relocate manifest due to collisions in output paths"))
			    })
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

					actual := readFileOk(filepath.Join(options.Output, "release.yaml"))
					expected := readFileOk("./fixtures/image_relocation/release_relocated.yaml")
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

					actual := readFileOk(options.Output)
					expected := readFileOk("./fixtures/image_relocation/release_relocated.yaml")
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

				actual := readFileOk(filepath.Join(options.Output, "release.yaml"))
				expected := readFileOk("./fixtures/image_relocation/release_relocated.yaml")
				Expect(actual).To(Equal(expected))
			})
		})
	})

	Describe("Flatteners", func() {
	    It("should preserve the normal manifest file name", func() {
	        for _, f := range flatteners {
	        	Expect(f("./manifest.yaml")).To(Equal("manifest.yaml"))
			}
	    })
	})

})

func readManifest(manifestPath string) *Manifest {
	manifest, err := NewManifest(manifestPath)
	Expect(err).NotTo(HaveOccurred())
	return manifest
}

func readFileOk(path string) string {
	content, err := ioutil.ReadFile(path)
	Expect(err).NotTo(HaveOccurred())
	return string(content)
}

func readURLOk(url string) string {
	t := &http.Transport{}
	t.RegisterProtocol("file", http.NewFileTransport(http.Dir("/")))
	c := &http.Client{Transport: t}

	resp, err := c.Get(url)
	Expect(err).NotTo(HaveOccurred())
	defer resp.Body.Close()

	content, err := ioutil.ReadAll(resp.Body)
	Expect(err).NotTo(HaveOccurred())

	return string(content)
}

func filePath(u string) string {
	parsed, err := url.Parse(u)
	Expect(err).NotTo(HaveOccurred())
	return parsed.Path
}