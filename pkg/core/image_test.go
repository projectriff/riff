package core

import (
	"errors"
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/projectriff/riff/pkg/fileutils/mocks"
	"github.com/projectriff/riff/pkg/resource"
	"io/ioutil"
	"os"
	"path/filepath"
)

var _ = Describe("RelocateImages", func() {

	var (
		client      ImageClient
		mockFutils  *mocks.Copier
		mockChecker *mocks.Checker
		options     RelocateImagesOptions
		err         error
		testErr     error
	)

	BeforeEach(func() {
		mockFutils = new(mocks.Copier)
		mockChecker = new(mocks.Checker)
		client = NewImageClient(nil, mockFutils, mockChecker, resource.ListImages, ioutil.Discard)
		options.Registry = "reg"
		options.RegistryUser = "user"
		testErr = errors.New("test error")
	})

	JustBeforeEach(func() {
		err = client.RelocateImages(options)
	})

	AfterEach(func() {
		mockFutils.AssertExpectations(GinkgoT())
	})

	Describe("manifest relocation", func() {
		Context("when there are no collisions in input file names", func() {
			AssertSuccess := func() {
				It("should write the relocated manifest to the output directory", func() {
					Expect(err).NotTo(HaveOccurred())

					// manifest should be unchanged since it contains just a filename
					actualManifest := readManifest(filepath.Join(options.Output, "manifest.yaml"))
					expectedManifest := readManifest("./fixtures/image_relocation/manifest.yaml")
					Expect(actualManifest).To(Equal(expectedManifest))

					actualYAML := readFileOk(filepath.Join(options.Output, "istio.yaml"))
					expectedYAML := readFileOk("./fixtures/image_relocation/istio_relocated.yaml")
					Expect(actualYAML).To(Equal(expectedYAML))

					actualYAML = readFileOk(filepath.Join(options.Output, "release.yaml"))
					expectedYAML = readFileOk("./fixtures/image_relocation/release_relocated.yaml")
					Expect(actualYAML).To(Equal(expectedYAML))

					actualYAML = readFileOk(filepath.Join(options.Output, "build.yaml"))
					expectedYAML = readFileOk("./fixtures/image_relocation/build_relocated.yaml")
					Expect(actualYAML).To(Equal(expectedYAML))
				})

				It("should write a relocated image manifest to the output directory", func() {
					Expect(err).NotTo(HaveOccurred())

					actualImageManifest := readImageManifest(filepath.Join(options.Output, "image-manifest.yaml"))
					expectedImageManifest := readImageManifest("./fixtures/image_relocation/image-manifest-relocated.yaml")
					Expect(actualImageManifest).To(Equal(expectedImageManifest))
				})
			}

			BeforeEach(func() {
				options.Manifest = "./fixtures/image_relocation/manifest.yaml"
				options.Images = "./fixtures/image_relocation/image-manifest.yaml"
			})

			Context("when the output directory already exists", func() {
				BeforeEach(func() {
					dir, err := ioutil.TempDir("", "image-relocation-test")
					Expect(err).NotTo(HaveOccurred())
					options.Output = dir

					mockFutils.On("Copy", dir, "fixtures/image_relocation/images").Return(nil)
				})

				AssertSuccess()
			})

			Context("when the output directory does not exist", func() {
				BeforeEach(func() {
					dir, err := ioutil.TempDir("", "image-relocation-test")
					Expect(err).NotTo(HaveOccurred())
					options.Output = filepath.Join(dir, "new")

					mockFutils.On("Copy", options.Output, "fixtures/image_relocation/images").Return(nil)
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

			Context("when copying the binary images returns an error", func() {
				BeforeEach(func() {
					dir, err := ioutil.TempDir("", "image-relocation-test")
					Expect(err).NotTo(HaveOccurred())
					options.Output = dir

					mockFutils.On("Copy", dir, "fixtures/image_relocation/images").Return(testErr)
				})

				It("should return the error", func() {
					Expect(err).To(MatchError(testErr))
				})
			})
		})

		Context("when there are collisions in input file names", func() {
			BeforeEach(func() {
				options.Manifest = "./fixtures/image_relocation/colliding/manifest.yaml"
				options.Images = "./fixtures/image_relocation/image-manifest.yaml"

				dir, err := ioutil.TempDir("", "image-relocation-test")
				Expect(err).NotTo(HaveOccurred())
				options.Output = dir
			})

			Context("when the provided flatteners prevent collisions", func() {
				BeforeEach(func() {
					mockFutils.On("Copy", options.Output, "fixtures/image_relocation/images").Return(nil)
				})

				It("should avoid unintended collisions in the output manifest", func() {
					Expect(err).NotTo(HaveOccurred())

					actualManifest := readManifest(filepath.Join(options.Output, "manifest.yaml"))
					expectedManifest := readManifest("./fixtures/image_relocation/colliding/manifest_with_collisions_relocated.yaml")
					Expect(actualManifest).To(Equal(expectedManifest))

					istios := actualManifest.Istio
					Expect(len(istios)).To(Equal(1))
					actualIstio := readFileOk(filepath.Join(options.Output, istios[0]))
					expectedIstio := readFileOk("./fixtures/image_relocation/istio_relocated.yaml")
					Expect(actualIstio).To(Equal(expectedIstio))

					knatives := actualManifest.Knative
					Expect(len(knatives)).To(Equal(1))
					actualBuild1 := readFileOk(filepath.Join(options.Output, knatives[0]))
					expectedBuild := readFileOk("./fixtures/image_relocation/build_relocated.yaml")
					Expect(actualBuild1).To(Equal(expectedBuild))

					namespaces := actualManifest.Namespace
					Expect(len(namespaces)).To(Equal(1))
					actualBuild2 := readFileOk(filepath.Join(options.Output, namespaces[0]))
					Expect(actualBuild2).To(Equal(expectedBuild))
				})
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

	Describe("resource file relocation", func() {
		Context("when the resource file is specified using a file path", func() {
			BeforeEach(func() {
				options.SingleFile = "./fixtures/image_relocation/release.yaml"
				options.Images = "./fixtures/image_relocation/image-manifest.yaml"
			})

			Context("when the output path is a directory", func() {
				BeforeEach(func() {
					dir, err := ioutil.TempDir("", "image-relocation-test")
					Expect(err).NotTo(HaveOccurred())
					options.Output = dir
				})

				It("should write the relocated resource file to the output directory", func() {
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

				It("should write the relocated resource file to the output path", func() {
					Expect(err).NotTo(HaveOccurred())

					actual := readFileOk(options.Output)
					expected := readFileOk("./fixtures/image_relocation/release_relocated.yaml")
					Expect(actual).To(Equal(expected))
				})
			})
		})

		Context("when the resource file is specified using a URL", func() {
			BeforeEach(func() {
				cwd, err := os.Getwd()
				Expect(err).NotTo(HaveOccurred())
				options.SingleFile = fmt.Sprintf("file://%s/fixtures/image_relocation/release.yaml", cwd) // local URL so test can run without network
				options.Images = "./fixtures/image_relocation/image-manifest.yaml"

				dir, err := ioutil.TempDir("", "image-relocation-test")
				Expect(err).NotTo(HaveOccurred())
				options.Output = dir
			})

			It("should write the relocated resource file to the output directory", func() {
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

	Describe("binary image copying", func() {
		Context("when there are no binary images", func() {
			BeforeEach(func() {
				options.Manifest = "./fixtures/image_relocation/manifest.yaml"
				options.Images = "./fixtures/image_relocation/no_binary_images/image-manifest.yaml"

				dir, err := ioutil.TempDir("", "image-relocation-test")
				Expect(err).NotTo(HaveOccurred())
				options.Output = dir
			})

			It("should succeed", func() {
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})

})

var _ = Describe("SystemDownload", func() {

	var (
		client      ImageClient
		mockCopier  *mocks.Copier
		mockChecker *mocks.Checker
		options     DownloadSystemOptions
		err         error
	)

	BeforeEach(func() {
		mockCopier = new(mocks.Copier)
		mockChecker = new(mocks.Checker)
		client = NewImageClient(nil, mockCopier, mockChecker, resource.ListImages, ioutil.Discard)
	})

	JustBeforeEach(func() {
		err = client.DownloadSystem(options)
	})

	AfterEach(func() {
		mockCopier.AssertExpectations(GinkgoT())
	})

	Describe("system download", func() {
		Context("when there are no collisions in input file names", func() {
			AssertSuccess := func() {
				It("should write the relocated manifest to the output directory", func() {
					Expect(err).NotTo(HaveOccurred())

					// manifest should be unchanged since it contains just a filename
					actualManifest := readManifest(filepath.Join(options.Output, "manifest.yaml"))
					expectedManifest := readManifest("./fixtures/image_relocation/manifest.yaml")
					Expect(actualManifest).To(Equal(expectedManifest))

					actualYAML := readFileOk(filepath.Join(options.Output, "istio.yaml"))
					expectedYAML := readFileOk("./fixtures/image_relocation/istio.yaml")
					Expect(actualYAML).To(Equal(expectedYAML))

					actualYAML = readFileOk(filepath.Join(options.Output, "release.yaml"))
					expectedYAML = readFileOk("./fixtures/image_relocation/release.yaml")
					Expect(actualYAML).To(Equal(expectedYAML))

					actualYAML = readFileOk(filepath.Join(options.Output, "build.yaml"))
					expectedYAML = readFileOk("./fixtures/image_relocation/build.yaml")
					Expect(actualYAML).To(Equal(expectedYAML))
				})
			}

			BeforeEach(func() {
				options.Manifest = "./fixtures/image_relocation/manifest.yaml"
			})

			Context("when the output directory already exists", func() {
				BeforeEach(func() {
					dir, err := ioutil.TempDir("", "system-download-test")
					Expect(err).NotTo(HaveOccurred())
					options.Output = dir
				})

				AssertSuccess()
			})

			Context("when the output directory does not exist", func() {
				BeforeEach(func() {
					dir, err := ioutil.TempDir("", "system-download-test")
					Expect(err).NotTo(HaveOccurred())
					options.Output = filepath.Join(dir, "new")
				})

				AssertSuccess()
			})

			Context("when the output directory is actually a file", func() {
				BeforeEach(func() {
					dir, err := ioutil.TempDir("", "system-download-test")
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

				dir, err := ioutil.TempDir("", "system-download-test")
				Expect(err).NotTo(HaveOccurred())
				options.Output = dir
			})

			Context("when the provided flatteners prevent collisions", func() {
				It("should avoid unintended collisions in the output manifest", func() {
					Expect(err).NotTo(HaveOccurred())

					actualManifest := readManifest(filepath.Join(options.Output, "manifest.yaml"))
					expectedManifest := readManifest("./fixtures/image_relocation/colliding/manifest_with_collisions_relocated.yaml")
					Expect(actualManifest).To(Equal(expectedManifest))

					istios := actualManifest.Istio
					Expect(len(istios)).To(Equal(1))
					actualIstio := readFileOk(filepath.Join(options.Output, istios[0]))
					expectedIstio := readFileOk("./fixtures/image_relocation/istio.yaml")
					Expect(actualIstio).To(Equal(expectedIstio))

					knatives := actualManifest.Knative
					Expect(len(knatives)).To(Equal(1))
					actualBuild1 := readFileOk(filepath.Join(options.Output, knatives[0]))
					expectedBuild := readFileOk("./fixtures/image_relocation/build.yaml")
					Expect(actualBuild1).To(Equal(expectedBuild))

					namespaces := actualManifest.Namespace
					Expect(len(namespaces)).To(Equal(1))
					actualBuild2 := readFileOk(filepath.Join(options.Output, namespaces[0]))
					Expect(actualBuild2).To(Equal(expectedBuild))
				})
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

func readImageManifest(imageManifestPath string) *ImageManifest {
	imageManifest, err := NewImageManifest(imageManifestPath)
	Expect(err).NotTo(HaveOccurred())
	return imageManifest
}
