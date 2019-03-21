/*
 * Copyright 2018 The original author or authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     https://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package crd_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/projectriff/riff/pkg/core"
	"path/filepath"
	"runtime"
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
				manifestPath = "./fixtures/manifest/invalid.yaml"
			})

			It("should return a suitable error", func() {
				Expect(err).To(MatchError(HavePrefix("Error parsing manifest file: ")))
			})
		})

		Context("when the manifest has the wrong version", func() {
			BeforeEach(func() {
				manifestPath = "./fixtures/manifest/wrongversion.yaml"
			})

			It("should return a suitable error", func() {
				Expect(err).To(MatchError("Manifest has unsupported version: 0.0"))
			})
		})

		Context("when the manifest does not specify the istio array", func() {
			BeforeEach(func() {
				manifestPath = "./fixtures/manifest/noistio.yaml"
			})

			It("should return a suitable error", func() {
				Expect(err).To(MatchError(HavePrefix("Manifest is incomplete: istio array missing: ")))
			})
		})

		Context("when the manifest does not specify the knative array", func() {
			BeforeEach(func() {
				manifestPath = "./fixtures/manifest/noknative.yaml"
			})

			It("should return a suitable error", func() {
				Expect(err).To(MatchError(HavePrefix("Manifest is incomplete: knative array missing: ")))
			})
		})

		Context("when the manifest does not specify the namespace array", func() {
			BeforeEach(func() {
				manifestPath = "./fixtures/manifest/nonamespace.yaml"
			})

			It("should return a suitable error", func() {
				Expect(err).To(MatchError(HavePrefix("Manifest is incomplete: namespace array missing: ")))
			})
		})

		Context("when the manifest contains a resource with an absolute path", func() {
			BeforeEach(func() {
				manifestPath = filepath.Join("fixtures", "manifest")
				if runtime.GOOS == "windows" {
					manifestPath = filepath.Join(manifestPath, "absolutepath.windows.yaml")
				} else {
					manifestPath = filepath.Join(manifestPath, "absolutepath.yaml")
				}
			})

			It("should return a suitable error", func() {
				Expect(err).To(MatchError(ContainSubstring("resources must use a http or https URL or a relative path: absolute path not supported: ")))
			})
		})

		Context("when the manifest contains a resource with unsupported URL scheme", func() {
			BeforeEach(func() {
				manifestPath = "./fixtures/manifest/invalidscheme.yaml"
			})

			It("should return a suitable error", func() {
				Expect(err).To(MatchError("resources must use a http or https URL or a relative path: scheme file not supported: file:///x"))
			})
		})

		Context("when the manifest is valid", func() {
			BeforeEach(func() {
				manifestPath = "./fixtures/manifest/valid.yaml"
			})

			It("should return with no error", func() {
				Expect(err).NotTo(HaveOccurred())
			})

			It("should parse the istio array", func() {
				Expect(manifest.Istio).To(ConsistOf("istio-crds", "http://istio-release"))
			})

			It("should parse the Knative array", func() {
				Expect(manifest.Knative).To(ConsistOf("build-release", "https://serving-release", "eventing-release"))
			})

			It("should parse the Knative array", func() {
				Expect(manifest.Namespace).To(ConsistOf("buildtemplate-release"))
			})
		})
	})

})
