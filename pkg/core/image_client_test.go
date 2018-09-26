/*
 * Copyright 2018 The original author or authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package core_test

import (
	"errors"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/projectriff/riff/pkg/core"
	"github.com/projectriff/riff/pkg/docker/mocks"
	"github.com/stretchr/testify/mock"
	"io/ioutil"
	"os"
	"path/filepath"
)

var _ = Describe("ImageClient", func() {
	var (
		imageClient core.ImageClient
		mockDocker  *mocks.Docker
		testError   error
	)

	BeforeEach(func() {
		mockDocker = new(mocks.Docker)
		testError = errors.New("test error")
	})

	JustBeforeEach(func() {
		imageClient = core.NewImageClient(mockDocker)
	})

	AfterEach(func() {
		mockDocker.AssertExpectations(GinkgoT())
	})

	Describe("LoadAndTagImages", func() {
		var (
			options core.LoadAndTagImagesOptions
			err     error
		)

		JustBeforeEach(func() {
			err = imageClient.LoadAndTagImages(options)
		})

		Context("when the manifest has a digest for each image", func() {
			BeforeEach(func() {
				options.Images = "fixtures/image_client/complete.yaml"
				mockDocker.On("LoadAndTagImage", "a/b", "1", "fixtures/image_client/images/1").Return(nil)
				mockDocker.On("LoadAndTagImage", "c/d", "2", "fixtures/image_client/images/2").Return(nil)
			})

			It("should succeed", func() {
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when the manifest has a missing digest", func() {
			BeforeEach(func() {
				options.Images = "fixtures/image_client/incomplete.yaml"
				mockDocker.On("LoadAndTagImage", "a/b", "1", "fixtures/image_client/images/1").Return(nil).Maybe()
			})

			It("should succeed", func() {
				Expect(err).To(MatchError("image manifest fixtures/image_client/incomplete.yaml does not specify a digest for image c/d"))
			})
		})

		Context("when the docker client returns an error", func() {
			BeforeEach(func() {
				mockDocker.On("LoadAndTagImage", mock.Anything, mock.Anything, mock.Anything).Return(testError).Once()
				options.Images = "fixtures/image_client/complete.yaml"
			})

			It("should propagate the error", func() {
				Expect(err).To(MatchError(testError))
			})
		})

		Context("when the image manifest cannot be read", func() {
			BeforeEach(func() {
				options.Images = "no/such"
			})

			It("should return a suitable error", func() {
				Expect(err).To(MatchError("error reading image manifest file: open no/such: no such file or directory"))
			})
		})
	})

	Describe("PushImages", func() {
		var (
			options core.PushImagesOptions
			err     error
		)

		JustBeforeEach(func() {
			err = imageClient.PushImages(options)
		})

		Context("when the manifest has a digest for each image", func() {
			BeforeEach(func() {
				options.Images = "fixtures/image_client/complete.yaml"
				mockDocker.On("LoadAndTagImage", mock.Anything, mock.Anything, mock.Anything).Return(nil).Twice()
				mockDocker.On("PushImage", "a/b").Return(nil)
				mockDocker.On("PushImage", "c/d").Return(nil)
			})

			It("should succeed", func() {
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when the manifest has a missing digest", func() {
			BeforeEach(func() {
				options.Images = "fixtures/image_client/incomplete.yaml"
				mockDocker.On("LoadAndTagImage", "a/b", "1", "fixtures/image_client/images/1").Return(nil).Maybe()
			})

			It("should succeed", func() {
				Expect(err).To(MatchError("image manifest fixtures/image_client/incomplete.yaml does not specify a digest for image c/d"))
			})
		})

		Context("when the docker client returns an error", func() {
			BeforeEach(func() {
				mockDocker.On("LoadAndTagImage", mock.Anything, mock.Anything, mock.Anything).Return(nil).Twice()
				mockDocker.On("PushImage", mock.Anything).Return(testError).Once()
				options.Images = "fixtures/image_client/complete.yaml"
			})

			It("should propagate the error", func() {
				Expect(err).To(MatchError(testError))
			})
		})

		Context("when the image manifest cannot be read", func() {
			BeforeEach(func() {
				options.Images = "no/such"
			})

			It("should return a suitable error", func() {
				Expect(err).To(MatchError("error reading image manifest file: open no/such: no such file or directory"))
			})
		})
	})

	Describe("PullImages", func() {
		var (
			options               core.PullImagesOptions
			workDir               string
			imagesDir             string
			err                   error
			expectedImageManifest *core.ImageManifest
		)

		BeforeEach(func() {
			// Ensure optional options do not leak from one test to another
			options.Output = ""
			options.ContinueOnMismatch = false

			// Avoid tests updating fixtures
			workDir, err = ioutil.TempDir("", "image_client_test")
			Expect(err).NotTo(HaveOccurred())
			options.Images = copyFile("fixtures/image_client/complete.yaml", workDir)
			imagesDir = filepath.Join(workDir, "images")
			expectedImageManifest = core.EmptyImageManifest()
			expectedImageManifest.Images["a/b"] = "1"
			expectedImageManifest.Images["c/d"] = "2"
		})

		AfterEach(func() {
			err = os.RemoveAll(workDir)
			Expect(err).NotTo(HaveOccurred())
		})

		JustBeforeEach(func() {
			err = imageClient.PullImages(options)
		})

		Context("when the returned digests match those in the manifest", func() {
			BeforeEach(func() {
				mockDocker.On("PullImage", "a/b", imagesDir).Return("1", nil)
				mockDocker.On("PullImage", "c/d", imagesDir).Return("2", nil)
			})

			It("should succeed", func() {
				Expect(err).NotTo(HaveOccurred())
			})

			It("should output the correct image manifest", func() {
				Expect(actualImageManifest(options.Images)).To(Equal(expectedImageManifest))
			})
		})

		Context("when the returned digests conflict with those in the manifest", func() {
			BeforeEach(func() {
				mockDocker.On("PullImage", "a/b", imagesDir).Return("1", nil).Maybe()
				mockDocker.On("PullImage", "c/d", imagesDir).Return("3", nil)
			})

			Context("when conflicts are not allowed", func() {
				It("should fail with a suitable error", func() {
					Expect(err).To(MatchError(`image "c/d" had digest 2 in the original manifest, but the pulled version has digest 3`))
				})
			})

			Context("when conflicts are allowed", func() {
				BeforeEach(func() {
					options.ContinueOnMismatch = true
					expectedImageManifest.Images["c/d"] = "3"
				})

				It("should succeed", func() {
					Expect(err).NotTo(HaveOccurred())
				})

				It("should output the correct image manifest", func() {
					Expect(actualImageManifest(options.Images)).To(Equal(expectedImageManifest))
				})
			})
		})

		Context("when output directory is specified", func() {
			BeforeEach(func() {
				options.Output = filepath.Join(workDir, "image_client_test_output")
				imagesOutputDir := filepath.Join(options.Output, "images")
				mockDocker.On("PullImage", "a/b", imagesOutputDir).Return("1", nil)
				mockDocker.On("PullImage", "c/d", imagesOutputDir).Return("2", nil)
			})

			It("should succeed", func() {
				Expect(err).NotTo(HaveOccurred())
			})

			It("should output the correct image manifest", func() {
				Expect(actualImageManifest(filepath.Join(options.Output, "image-manifest.yaml"))).To(Equal(expectedImageManifest))
			})
		})

		Context("when output images directory cannot be created", func() {
			BeforeEach(func() {
				options.Output = filepath.Join(workDir, "image_client_test_output")
				err = ioutil.WriteFile(options.Output, []byte{1}, 0644)
				Expect(err).NotTo(HaveOccurred())
			})

			It("should return a suitable error", func() {
				Expect(err).To(HaveOccurred())
				if _, ok := err.(*os.PathError); !ok {
					Fail("Unexpected error type")
				}
			})
		})

		Context("when the docker client returns an error", func() {
			BeforeEach(func() {
				mockDocker.On("PullImage", mock.Anything, mock.Anything).Return("", testError).Once()
			})

			It("should propagate the error", func() {
				Expect(err).To(MatchError(testError))
			})
		})

		Context("when the image manifest cannot be read", func() {
			BeforeEach(func() {
				options.Images = "no/such"
			})

			It("should return a suitable error", func() {
				Expect(err).To(MatchError("error reading image manifest file: open no/such: no such file or directory"))
			})
		})
	})
})

func copyFile(src string, destDir string) string {
	contents, err := ioutil.ReadFile(src)
	Expect(err).NotTo(HaveOccurred())
	dest := filepath.Join(destDir, filepath.Base(src))
	err = ioutil.WriteFile(dest, contents, 0644)
	Expect(err).NotTo(HaveOccurred())
	return dest
}

func actualImageManifest(path string) *core.ImageManifest {
	m, err := core.NewImageManifest(path)
	Expect(err).NotTo(HaveOccurred())
	return m
}
