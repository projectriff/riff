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
 *
 */

package resource_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/projectriff/riff/pkg/resource"
	"os"
	"path/filepath"
)

var _ = Describe("ListImages", func() {
	var (
		res     string
		baseDir string
		images  []string
		err     error
	)

	BeforeEach(func() {
		wd, err := os.Getwd()
		Expect(err).NotTo(HaveOccurred())
		baseDir = filepath.Join(wd, "fixtures")
	})

	JustBeforeEach(func() {
		images, err = resource.ListImages(res, baseDir)
	})

	Context("when the resource file names an image directly", func() {
		BeforeEach(func() {
			res = "simple.yaml"
		})

		It("should list the image", func() {
			Expect(err).NotTo(HaveOccurred())
			Expect(images).To(ConsistOf("gcr.io/knative-releases/x/y"))
		})
	})

	Context("when the resource file names images as arguments", func() {
		BeforeEach(func() {
			res = "arg.yaml"
		})

		It("should list the images", func() {
			Expect(err).NotTo(HaveOccurred())
			Expect(images).To(ConsistOf("gcr.io/knative-releases/a/b", "gcr.io/knative-releases/c/d"))
		})
	})

	Context("when the resource file names an image using a key ending in 'Image'", func() {
		BeforeEach(func() {
			res = "suffix.yaml"
		})

		It("should list the images", func() {
			Expect(err).NotTo(HaveOccurred())
			Expect(images).To(ConsistOf("gcr.io/knative-releases/e/f", "k8s.gcr.io/fluentd-elasticsearch:v2.0.4"))
		})
	})


	Context("when the resource file contains block scalars containing images", func() {
		BeforeEach(func() {
			res = "block.yaml"
		})

		It("should list the images", func() {
			Expect(err).NotTo(HaveOccurred())
			Expect(images).To(ConsistOf("docker.io/istio/proxy_init:1.0.1"))
		})
	})

	Context("when using a realistic resource file", func() {
		BeforeEach(func() {
			res = "complex.yaml"
		})

		It("should list the images in the resource file", func() {
			Expect(err).NotTo(HaveOccurred())
			Expect(images).To(ConsistOf("gcr.io/knative-releases/github.com/knative/build/cmd/controller@sha256:3981b19105aabf3ed66db38c15407dc7accf026f4f4703d7e0ca7986ffd37d99",
				"gcr.io/knative-releases/github.com/knative/build/cmd/creds-init@sha256:b5dff24742c5c8ac4673dc991e3f960d11b58efdf751d26c54ec5144c48eef30",
				"gcr.io/knative-releases/github.com/knative/build/cmd/git-init@sha256:fe0d19e5da3fc9e7da20abc13d032beafcc283358a8325188dced62536a66e54",
				"gcr.io/knative-releases/github.com/knative/build/cmd/webhook@sha256:b9a97b7d360e10e540edfc9329e4f1c01832e58bf57d5dddea5c3a664f64bfc6",
				"docker.io/istio/proxyv2:0.8.0",
				"gcr.io/knative-releases/github.com/knative/serving/cmd/activator@sha256:e83258dd5858c8b1e92dbd413d0857ad2b22a7c4215ed911f256f68e2972f362",
				"gcr.io/knative-releases/github.com/knative/serving/cmd/autoscaler@sha256:76222399addc02454db9837ea3ff54bae29849168586051a9d0180daa2c1a805",
				"gcr.io/knative-releases/github.com/knative/serving/cmd/queue@sha256:99c841aa72c2928d1bf333348a848b5afb182715a2a0441da6282c86d4be807e",
				"k8s.gcr.io/fluentd-elasticsearch:v2.0.4",
				"gcr.io/knative-releases/github.com/knative/serving/cmd/controller@sha256:28db335f18cbd2a015fd218b9c7ce30b4366898fa3728a7f6dab6537991de028",
				"gcr.io/knative-releases/github.com/knative/serving/cmd/webhook@sha256:50ea89c48f8890fbe0cee336fc5cbdadcfe6884afbe5977db5d66892095b397d",
			))
		})
	})

	Context("when the resource file is not found", func() {
	    BeforeEach(func() {
	        res = "nosuch.yaml"
	    })

	    It("should return a suitable error", func() {
	        Expect(err).To(MatchError(HaveSuffix("no such file or directory")))
	    })
	})

	Context("when the resource file contains invalid YAML", func() {
	    BeforeEach(func() {
	        res = "invalid.yaml"
	    })

	    It("should return a suitable error", func() {
	        Expect(err).To(MatchError(HavePrefix("error parsing resource file")))
	    })
	})
})
