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

package core

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("imageMapper", func() {
	const (
		testImage = "gcr.io/knative-releases/github.com/knative/build/cmd/creds-init@shannn:deadbeef"
		testYaml  = `image: "` + testImage + `"
image: ` + testImage + `
`
		testRegistry = "testregistry.com"
		testUser     = "testuser"
		mappedImage  = "testregistry.com/testuser/github.com/knative/build/cmd/creds-init@shannn:deadbeef"
		mappedYaml   = `image: "` + mappedImage + `"
image: ` + mappedImage + `
`
	)
	var (
		registry string
		user     string
		images   []imageName
		mapper   *imageMapper
		err      error
		input    []byte
		output   []byte
	)

	BeforeEach(func() {
		registry = testRegistry
		user = testUser
	})

	JustBeforeEach(func() {
		mapper, err = newImageMapper(registry, user, images)
	})

	Describe("newImageMapper", func() {
		Context("when the registry name contains a forward slash", func() {
			BeforeEach(func() {
				registry = "/"
			})

			It("should return a suitable error", func() {
				Expect(err).To(MatchError("invalid registry hostname: '/' contains '/'"))
			})
		})

		Context("when the registry name contains a double quote", func() {
			BeforeEach(func() {
				registry = `"`
			})

			It("should return a suitable error", func() {
				Expect(err).To(MatchError(`invalid registry hostname: '"' contains '"'`))
			})
		})

		Context("when the registry name contains a space", func() {
			BeforeEach(func() {
				registry = " "
			})

			It("should return a suitable error", func() {
				Expect(err).To(MatchError("invalid registry hostname: ' ' contains ' '"))
			})
		})

		Context("when the user contains a forward slash", func() {
			BeforeEach(func() {
				user = "/"
			})

			It("should return a suitable error", func() {
				Expect(err).To(MatchError("invalid user: '/' contains '/'"))
			})
		})

		Context("when the user contains a double quote", func() {
			BeforeEach(func() {
				user = `"`
			})

			It("should return a suitable error", func() {
				Expect(err).To(MatchError(`invalid user: '"' contains '"'`))
			})
		})

		Context("when the user contains a space", func() {
			BeforeEach(func() {
				user = " "
			})

			It("should return a suitable error", func() {
				Expect(err).To(MatchError("invalid user: ' ' contains ' '"))
			})
		})

		Context("when an image contains no user", func() {
			BeforeEach(func() {
				images = []imageName{"a"}
			})

			It("should return a suitable error", func() {
				Expect(err).To(MatchError("invalid image: user missing: a"))
			})
		})

		Context("when an image omits the host", func() {
			BeforeEach(func() {
				images = []imageName{"a/b"}
			})

			It("should default the host", func() {
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when an image contains a double quote", func() {
			BeforeEach(func() {
				images = []imageName{`"`}
			})

			It("should return a suitable error", func() {
				Expect(err).To(MatchError(`invalid image: '"' contains '"'`))
			})
		})

		Context("when an image contains a space", func() {
			BeforeEach(func() {
				images = []imageName{" "}
			})

			It("should return a suitable error", func() {
				Expect(err).To(MatchError("invalid image: ' ' contains ' '"))
			})
		})
	})

	Describe("mapImages", func() {
		BeforeEach(func() {
			input = []byte(testYaml)
		})

		JustBeforeEach(func() {
			Expect(err).NotTo(HaveOccurred())
			output = mapper.mapImages(input)
		})

		Context("when the list of images is empty", func() {
			BeforeEach(func() {
				images = []imageName{}
			})

			It("should not perform any mappings", func() {
				Expect(output).To(Equal([]byte(testYaml)))
			})
		})

		Context("when the list of images is non-empty", func() {
			BeforeEach(func() {
				images = []imageName{testImage}
			})

			It("should perform the mappings", func() {
				Expect(string(output)).To(Equal(string([]byte(mappedYaml))))
			})
		})

		Describe("DockerHub special cases", func() {
			const (
				elidedImage   = "istio/sidecar_injector"
				explicitImage = "docker.io/istio/sidecar_injector"
				fullImage     = "index.docker.io/istio/sidecar_injector"
			)

			Context("when an image omits the docker.io host", func() {
				BeforeEach(func() {
					images = []imageName{elidedImage}
					input = []byte(`image: "` + elidedImage + `"
image: "` + explicitImage + `"
image: "` + fullImage + `"`)
				})

				It("should perform the mapping", func() {
					mappedImage := fmt.Sprintf("testregistry.com/testuser/sidecar_injector")
					Expect(string(output)).To(Equal(string([]byte(fmt.Sprintf(`image: %q
image: %q
image: %q`, mappedImage, mappedImage, mappedImage)))))
				})
			})

			Context("when an image includes the docker.io host", func() {
				BeforeEach(func() {
					images = []imageName{explicitImage}
					input = []byte(`image: "` + elidedImage + `"
image: "` + explicitImage + `"
image: "` + fullImage + `"`)
				})

				It("should perform the mapping", func() {
					mappedImage := fmt.Sprintf("testregistry.com/testuser/sidecar_injector")
					Expect(string(output)).To(Equal(string([]byte(fmt.Sprintf(`image: %q
image: %q
image: %q`, mappedImage, mappedImage, mappedImage)))))
				})
			})

			Context("when an image includes the index.docker.io host", func() {
				BeforeEach(func() {
					images = []imageName{fullImage}
					input = []byte(`image: "` + elidedImage + `"
image: "` + explicitImage + `"
image: "` + fullImage + `"`)
				})

				It("should perform the mapping", func() {
					mappedImage := fmt.Sprintf("testregistry.com/testuser/sidecar_injector")
					Expect(string(output)).To(Equal(string([]byte(fmt.Sprintf(`image: %q
image: %q
image: %q`, mappedImage, mappedImage, mappedImage)))))
				})
			})
		})

		Describe("detailed parsing", func() {
			Context("when an image is wrapped in double quotes", func() {
				BeforeEach(func() {
					images = []imageName{"x/y/z"}
					input = []byte(`"x/y/z"`)
					registry = "r"
					user = "u"
				})

				It("should map the image", func() {
					Expect(string(output)).To(Equal(string([]byte(`"r/u/z"`))))
				})
			})

			Context("when an image is preceded by a space", func() {
				Context("when an image is followed by a space", func() {
					BeforeEach(func() {
						images = []imageName{"x/y/z"}
						input = []byte(" x/y/z ")
						registry = "r"
						user = "u"
					})

					It("should map the image", func() {
						Expect(string(output)).To(Equal(string([]byte(" r/u/z "))))
					})
				})

				Context("when an image is followed by a newline", func() {
					BeforeEach(func() {
						images = []imageName{"x/y/z"}
						input = []byte(" x/y/z\n")
						registry = "r"
						user = "u"
					})

					It("should map the image", func() {
						Expect(string(output)).To(Equal(string([]byte(" r/u/z\n"))))
					})
				})

				Context("when an image is at the end of the string", func() {
					BeforeEach(func() {
						images = []imageName{"x/y/z"}
						input = []byte(" x/y/z")
						registry = "r"
						user = "u"
					})

					It("should map the image", func() {
						Expect(string(output)).To(Equal(string([]byte(" r/u/z"))))
					})
				})
			})
		})

		Describe("edge cases", func() {
			Context("when one unmapped image is a prefix of another", func() {
				Context("when the image is wrapped in double quotes", func() {
					BeforeEach(func() {
						images = []imageName{"x/y/z", "x/y/z/a"}
						input = []byte(`image: "x/y/z/a"`)
						registry = "r"
						user = "u"
					})

					It("should map the image rather than the substring", func() {
						Expect(string(output)).To(Equal(string([]byte(`image: "r/u/z/a"`))))
					})
				})

				Context("when the image is preceded by a space", func() {
					BeforeEach(func() {
						images = []imageName{"x/y/z", "x/y/z/a"}
						input = []byte("image: x/y/z/a")
						registry = "r"
						user = "u"
					})

					It("should map the image rather than the substring", func() {
						Expect(string(output)).To(Equal(string([]byte("image: r/u/z/a"))))
					})
				})
			})

			Context("when an unmapped image is a prefix of an image in the string", func() {
				Context("when the image is wrapped in double quotes", func() {
					BeforeEach(func() {
						images = []imageName{"x/y/z"}
						input = []byte(`image: "x/y/z/a"`)
						registry = "r"
						user = "u"
					})

					It("does not map the prefix", func() {
						Expect(string(output)).To(Equal(string([]byte(`image: "x/y/z/a"`))))
					})
				})

				Context("when the image is preceded by a space", func() {
					BeforeEach(func() {
						images = []imageName{"x/y/z"}
						input = []byte("image: x/y/z/a")
						registry = "r"
						user = "u"
					})

					It("maps the substring (a known limitation)", func() {
						Expect(string(output)).To(Equal(string([]byte("image: r/u/z/a"))))
					})
				})
			})

			Context("when what looks like an image is not preceeded by a double quote or space", func() {
				BeforeEach(func() {
					images = []imageName{"x/y/z"}
					input = []byte("http://x/y/z")
					registry = "r"
					user = "u"
				})

				It("should not map the apparent image", func() {
					Expect(string(output)).To(Equal(string([]byte("http://x/y/z"))))
				})
			})
		})
	})

})
