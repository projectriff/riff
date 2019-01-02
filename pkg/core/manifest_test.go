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
	"os"
	"path/filepath"
	"runtime"

	"github.com/projectriff/riff/pkg/test_support"

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
				Expect(err).To(MatchError(HavePrefix("manifest is incomplete: istio array missing: ")))
			})
		})

		Context("when the manifest does not specify the knative array", func() {
			BeforeEach(func() {
				manifestPath = "./fixtures/manifest/noknative.yaml"
			})

			It("should return a suitable error", func() {
				Expect(err).To(MatchError(HavePrefix("manifest is incomplete: knative array missing: ")))
			})
		})

		Context("when the manifest does not specify the namespace array", func() {
			BeforeEach(func() {
				manifestPath = "./fixtures/manifest/nonamespace.yaml"
			})

			It("should return a suitable error", func() {
				Expect(err).To(MatchError(HavePrefix("manifest is incomplete: namespace array missing: ")))
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

			It("should parse the build template array", func() {
				Expect(manifest.Namespace).To(ConsistOf("buildtemplate-release"))
			})

			Describe("ResourceAbsolutePath", func() {
				var manifestDir string

				BeforeEach(func() {
					wd, err := os.Getwd()
					Expect(err).NotTo(HaveOccurred())
					manifestDir = filepath.Join(wd, "fixtures", "manifest")
				})

				It("should return a http URL unchanged", func() {
					Expect(manifest.ResourceAbsolutePath("http://istio-release")).To(Equal("http://istio-release"))
				})

				It("should return a https URL unchanged", func() {
					Expect(manifest.ResourceAbsolutePath("https://serving-release")).To(Equal("https://serving-release"))
				})

				It("should return the absolute file path from a file URL", func() {
					wd, err := os.Getwd()
					Expect(err).NotTo(HaveOccurred())
					Expect(manifest.ResourceAbsolutePath(test_support.FileURL(wd))).To(Equal(wd))
				})

				It("should return the absolute equivalent of a relative path", func() {
					Expect(manifest.ResourceAbsolutePath("file")).To(Equal(filepath.Join(manifestDir, "file")))
				})

				It("should return an absolute file path unchanged", func() {
					path := filepath.Join(manifestDir, "file")
					Expect(manifest.ResourceAbsolutePath(path)).To(Equal(path))
				})
			})
		})
	})

	Describe("ResolveManifest", func() {
		var (
			manifests    map[string]*core.Manifest
			testManifest *core.Manifest
			manifest     *core.Manifest
			err          error
		)

		BeforeEach(func() {
			testManifest = &core.Manifest{}
			manifests = map[string]*core.Manifest{"test": testManifest}
		})

		JustBeforeEach(func() {
			manifest, err = core.ResolveManifest(manifests, "test")
		})

		It("should return with no error", func() {
			Expect(err).NotTo(HaveOccurred())
		})

		It("should resolve the manifest correctly", func() {
			Expect(manifest).To(Equal(testManifest))
		})

		Describe("ResourceAbsolutePath", func() {
			It("should return a suitable error", func() {
				_, err := manifest.ResourceAbsolutePath("file")
				Expect(err).To(MatchError("relative path undefined since manifest was not read from a directory"))
			})
		})
	})
})
