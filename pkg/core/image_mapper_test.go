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

	"github.com/projectriff/riff/pkg/image"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("imageMapper", func() {
	const (
		testImage = "some.registry.com/some-user/some/path"
		testYaml  = `image: "` + testImage + `"
image: ` + testImage + `
`
		testRegistry         = "testregistry.com"
		testUser             = "testuser"
		mappedImage          = "testregistry.com/testuser/some/path"
		flattenedMappedImage = "testregistry.com/testuser/some-user-some-path-3236c106420c1d0898246e1d2b6ba8b6"
		mappedYaml           = `image: "` + mappedImage + `"
image: ` + mappedImage + `
`
		flattenedMappedYaml = `image: "` + flattenedMappedImage + `"
image: ` + flattenedMappedImage + `
`
	)
	var (
		registry string
		user     string
		images   []image.Name
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

		Context("when an image omits the host", func() {
			BeforeEach(func() {
				images = imageNames("a/b")
			})

			It("should default the host", func() {
				Expect(err).NotTo(HaveOccurred())
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
				images = imageNames()
			})

			It("should not perform any mappings", func() {
				Expect(output).To(Equal([]byte(testYaml)))
			})
		})

		Context("when the list of images is non-empty", func() {
			BeforeEach(func() {
				images = imageNames(testImage)
			})

			Context("when the target host is local only", func() {
				const (
					mappedImage = "dev.local/testuser/some-user-some-path-3236c106420c1d0898246e1d2b6ba8b6:local"
					mappedYaml  = `image: "` + mappedImage + `"
image: ` + mappedImage + `
`
				)

				BeforeEach(func() {
					registry = "dev.local"
				})

				It("should perform the mappings using flattened image names", func() {
					Expect(string(output)).To(Equal(string([]byte(mappedYaml))))
				})
			})

			It("should perform the mappings using flattened image names", func() {
				Expect(string(output)).To(Equal(string([]byte(flattenedMappedYaml))))
			})

			Context("when the target host is local only", func() {
				const (
					flattenedMappedLocalImage = "dev.local/testuser/some-user-some-path-3236c106420c1d0898246e1d2b6ba8b6:local"
					flattenedMappedLocalYaml  = `image: "` + flattenedMappedLocalImage + `"
image: ` + flattenedMappedLocalImage + `
`
				)

				BeforeEach(func() {
					registry = "dev.local"
				})

				It("should perform the mappings using flattened image names", func() {
					Expect(string(output)).To(Equal(string([]byte(flattenedMappedLocalYaml))))
				})
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
					images = imageNames(elidedImage)
					input = []byte(`image: "` + elidedImage + `"
image: "` + explicitImage + `"
image: "` + fullImage + `"`)
				})

				It("should perform the mapping", func() {
					mappedImage := fmt.Sprintf("testregistry.com/testuser/istio-sidecar_injector-6bad934e7f077e63cae18277203bb414")
					Expect(string(output)).To(Equal(string([]byte(fmt.Sprintf(`image: %q
image: %q
image: %q`, mappedImage, mappedImage, mappedImage)))))
				})
			})

			Context("when an image includes the docker.io host", func() {
				BeforeEach(func() {
					images = imageNames(explicitImage)
					input = []byte(`image: "` + elidedImage + `"
image: "` + explicitImage + `"
image: "` + fullImage + `"`)
				})

				It("should perform the mapping", func() {
					mappedImage := fmt.Sprintf("testregistry.com/testuser/istio-sidecar_injector-6bad934e7f077e63cae18277203bb414")
					Expect(string(output)).To(Equal(string([]byte(fmt.Sprintf(`image: %q
image: %q
image: %q`, mappedImage, mappedImage, mappedImage)))))
				})
			})

			Context("when an image includes the index.docker.io host", func() {
				BeforeEach(func() {
					images = imageNames(fullImage)
					input = []byte(`image: "` + elidedImage + `"
image: "` + explicitImage + `"
image: "` + fullImage + `"`)
				})

				It("should perform the mapping", func() {
					mappedImage := fmt.Sprintf("testregistry.com/testuser/istio-sidecar_injector-6bad934e7f077e63cae18277203bb414")
					Expect(string(output)).To(Equal(string([]byte(fmt.Sprintf(`image: %q
image: %q
image: %q`, mappedImage, mappedImage, mappedImage)))))
				})
			})
		})

		Describe("ko image special cases", func() {
			const (
				koImage = "gcr.io/knative-releases/github.com/knative/build/cmd/creds-init@sha256:deadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef"
			)

			Context("when an image is named by ko", func() {
				BeforeEach(func() {
					images = imageNames(koImage)
					input = []byte(`image: "` + koImage + `"`)
				})

				It("should perform the mapping", func() {
					mappedImage := fmt.Sprintf("testregistry.com/testuser/knative-releases-github.com-knative-build-cmd-creds-init-b692cdd35af41412d71a4bc138128dd6") // omits the "@sha256:" piece
					Expect(string(output)).To(Equal(string([]byte(fmt.Sprintf(`image: %q`, mappedImage)))))
				})
			})
		})

		Describe("short path special cases", func() {
			const (
				shortPathImage = "k8s.gcr.io/fluentd-elasticsearch:v2.0.4"
			)

			Context("when an image path consists only of a single element", func() {
				BeforeEach(func() {
					images = imageNames(shortPathImage)
					input = []byte(`image: "` + shortPathImage + `"`)
				})

				It("should perform the mapping", func() {
					mappedImage := "testregistry.com/testuser/fluentd-elasticsearch-e71ae1e56d9a15bbcf26be734d13b608:v2.0.4"
					Expect(string(output)).To(Equal(string([]byte(fmt.Sprintf(`image: %q`, mappedImage)))))
				})
			})
		})

		Describe("detailed parsing", func() {
			Context("when an image is wrapped in double quotes", func() {
				BeforeEach(func() {
					images = imageNames("x.x/y/z")
					input = []byte(`"x.x/y/z"`)
					registry = "r.r"
					user = "u"
				})

				It("should map the image", func() {
					Expect(string(output)).To(Equal(string([]byte(`"r.r/u/y-z-622eedc03bbe568ed522e1e4903704b2"`))))
				})
			})

			Context("when an image is preceded by a space", func() {
				Context("when an image is followed by a space", func() {
					BeforeEach(func() {
						images = imageNames("x.x/y/z")
						input = []byte(" x.x/y/z ")
						registry = "r.r"
						user = "u"
					})

					It("should map the image", func() {
						Expect(string(output)).To(Equal(string([]byte(" r.r/u/y-z-622eedc03bbe568ed522e1e4903704b2 "))))
					})
				})

				Context("when an image is followed by a newline", func() {
					BeforeEach(func() {
						images = imageNames("x.x/y/z")
						input = []byte(" x.x/y/z\n")
						registry = "r.r"
						user = "u"
					})

					It("should map the image", func() {
						Expect(string(output)).To(Equal(string([]byte(" r.r/u/y-z-622eedc03bbe568ed522e1e4903704b2\n"))))
					})
				})

				Context("when an image is at the end of the string", func() {
					BeforeEach(func() {
						images = imageNames("x.x/y/z")
						input = []byte(" x.x/y/z")
						registry = "r.r"
						user = "u"
					})

					It("should map the image", func() {
						Expect(string(output)).To(Equal(string([]byte(" r.r/u/y-z-622eedc03bbe568ed522e1e4903704b2"))))
					})
				})
			})
		})

		Describe("edge cases", func() {
			Context("when one unmapped image is a prefix of another", func() {
				Context("when the image is wrapped in double quotes", func() {
					BeforeEach(func() {
						images = imageNames("x.x/y/z", "x.x/y/z/a")
						input = []byte(`image: "x.x/y/z/a"`)
						registry = "r.r"
						user = "u"
					})

					It("should map the image rather than the substring", func() {
						Expect(string(output)).To(Equal(string([]byte(`image: "r.r/u/y-z-a-ab71dcefbbc95badf34dcb58bc9afceb"`))))
					})
				})

				Context("when the image is preceded by a space", func() {
					BeforeEach(func() {
						images = imageNames("x.x/y/z", "x.x/y/z/a")
						input = []byte("image: x.x/y/z/a")
						registry = "r.r"
						user = "u"
					})

					It("should map the image rather than the substring", func() {
						Expect(string(output)).To(Equal(string([]byte("image: r.r/u/y-z-a-ab71dcefbbc95badf34dcb58bc9afceb"))))
					})
				})
			})

			Context("when an unmapped image is a prefix of an image in the string", func() {
				Context("when the image is wrapped in double quotes", func() {
					BeforeEach(func() {
						images = imageNames("x.x/y/z")
						input = []byte(`image: "x.x/y/z/a"`)
						registry = "r.r"
						user = "u"
					})

					It("does not map the prefix", func() {
						Expect(string(output)).To(Equal(string([]byte(`image: "x.x/y/z/a"`))))
					})
				})

				Context("when the image is preceded by a space", func() {
					BeforeEach(func() {
						images = imageNames("x.x/y/z")
						input = []byte("image: x.x/y/z/a")
						registry = "r.r"
						user = "u"
					})

					It("maps the substring (a known limitation)", func() {
						Expect(string(output)).To(Equal(string([]byte("image: r.r/u/y-z-622eedc03bbe568ed522e1e4903704b2/a"))))
					})
				})
			})

			Context("when what looks like an image is not preceeded by a double quote or space", func() {
				BeforeEach(func() {
					images = imageNames("x.x/y/z")
					input = []byte("http://x.x/y/z")
					registry = "r.r"
					user = "u"
				})

				It("should not map the apparent image", func() {
					Expect(string(output)).To(Equal(string([]byte("http://x.x/y/z"))))
				})
			})
		})
	})
})

func imageNames(names ...string) []image.Name {
	in := []image.Name{}
	for _, i := range names {
		in = append(in, parseImageNameOk(i))
	}
	return in
}

func parseImageNameOk(i string) image.Name {
	in, err := image.NewName(i)
	Expect(err).NotTo(HaveOccurred())
	return in
}
