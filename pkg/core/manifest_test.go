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

package core_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/projectriff/riff/pkg/core"
)

var _ = Describe("Manifest", func() {
	Describe("NewManifest", func() {

		var (
			manifestPath string
			manifest     *core.Manifest
			err          error
		)

		JustBeforeEach(func() {
			manifest, err = core.NewManifest(manifestPath)
		})

		Context("when an invalid path is provided", func() {
			BeforeEach(func() {
				manifestPath = ""
			})

			It("should return a suitable error", func() {
				Expect(err).To(MatchError(HavePrefix("Error reading manifest file: ")))
			})
		})

		Context("when the file contains invalid YAML", func() {
			BeforeEach(func() {
				manifestPath = "./fixtures/invalid.yaml"
			})

			It("should return a suitable error", func() {
				Expect(err).To(MatchError(HavePrefix("Error parsing manifest file: ")))
			})
		})

		Context("when the manifest has the wrong version", func() {
			BeforeEach(func() {
				manifestPath = "./fixtures/wrongversion.yaml"
			})

			It("should return a suitable error", func() {
				Expect(err).To(MatchError("Manifest has unsupported version: 0.0"))
			})
		})

		Context("when the manifest does not specify the istio array", func() {
			BeforeEach(func() {
				manifestPath = "./fixtures/noistio.yaml"
			})

			It("should return a suitable error", func() {
				Expect(err).To(MatchError(HavePrefix("Manifest is incomplete: istio array missing: ")))
			})
		})

		Context("when the manifest does not specify the knative array", func() {
			BeforeEach(func() {
				manifestPath = "./fixtures/noknative.yaml"
			})

			It("should return a suitable error", func() {
				Expect(err).To(MatchError(HavePrefix("Manifest is incomplete: knative array missing: ")))
			})
		})

		Context("when the manifest does not specify the namespace array", func() {
			BeforeEach(func() {
				manifestPath = "./fixtures/nonamespace.yaml"
			})

			It("should return a suitable error", func() {
				Expect(err).To(MatchError(HavePrefix("Manifest is incomplete: namespace array missing: ")))
			})
		})

		Context("when the manifest is valid", func() {
			BeforeEach(func() {
				manifestPath = "./fixtures/valid.yaml"
			})

			It("should return with no error", func() {
				Expect(err).NotTo(HaveOccurred())
			})

			It("should parse the istio array", func() {
				Expect(manifest.Istio).To(ConsistOf("istio-crds", "istio-release"))
			})

			It("should parse the Knative array", func() {
				Expect(manifest.Knative).To(ConsistOf("serving-release", "eventing-release", "stub-bus-release"))
			})

			It("should parse the Knative array", func() {
				Expect(manifest.Namespace).To(ConsistOf("buildtemplate-release"))
			})
		})
	})

})
