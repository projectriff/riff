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
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ImageManifest", func() {
	Describe("ImageNewManifest", func() {

		var (
			manifestPath string
			manifest     *ImageManifest
			err          error
		)

		JustBeforeEach(func() {
			manifest, err = NewImageManifest(manifestPath)
		})

		Context("when an invalid path is provided", func() {
			BeforeEach(func() {
				manifestPath = ""
			})

			It("should return a suitable error", func() {
				Expect(err).To(MatchError(HavePrefix("error reading image manifest file: ")))
			})
		})

		Context("when the file contains invalid YAML", func() {
			BeforeEach(func() {
				manifestPath = "./fixtures/image_manifest/invalid.yaml"
			})

			It("should return a suitable error", func() {
				Expect(err).To(MatchError(HavePrefix("error parsing image manifest file: ")))
			})
		})

		Context("when the image manifest has the wrong version", func() {
			BeforeEach(func() {
				manifestPath = "./fixtures/image_manifest/wrongversion.yaml"
			})

			It("should return a suitable error", func() {
				Expect(err).To(MatchError("image manifest has unsupported version: 0.0"))
			})
		})

		Context("when the image manifest does not specify an image map", func() {
			BeforeEach(func() {
				manifestPath = "./fixtures/image_manifest/noimages.yaml"
			})

			It("should return a suitable error", func() {
				Expect(err).To(MatchError(HavePrefix("image manifest is incomplete: images map is missing: ")))
			})
		})

		Context("when the manifest is valid", func() {
			BeforeEach(func() {
				manifestPath = "./fixtures/image_manifest/valid.yaml"
			})

			It("should return with no error", func() {
				Expect(err).NotTo(HaveOccurred())
			})

			It("should parse the images array", func() {
				Expect(manifest.Images).To(Equal(map[imageName]imageDigest{"gcr.io/cf-spring-funkytown/github.com/knative/serving/cmd/queue": "",
					"istio/sidecar_injector": "0123",
					"gcr.io/knative-releases/github.com/knative/eventing/cmd/controller@sha256:367a7a22bc689b794c38fc488b8774a94515727a2c12f2347622e6c40fe9c1e8": "456"}))
			})
		})
	})

})
