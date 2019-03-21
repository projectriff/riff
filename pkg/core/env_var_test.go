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

package core_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/projectriff/riff/pkg/core"
	"k8s.io/api/core/v1"
)

var _ = Describe("Environment variable parsing", func() {

	var (
		input  []string
		output []v1.EnvVar
		err    error
	)

	Describe("ParseEnvVar", func() {
		JustBeforeEach(func() {
			output, err = core.ParseEnvVar(input)
		})

		Context("when key not specified", func() {
			BeforeEach(func() {
				input = []string{"="}
			})

			It("should fail with a suitable error", func() {
				Expect(err).To(MatchError("unable to parse '=', the key part is missing"))
			})
		})

		Context("when equal sign not specified", func() {
			BeforeEach(func() {
				input = []string{"FOO:BAR"}
			})

			It("should fail with a suitable error", func() {
				Expect(err).To(MatchError("unable to parse 'FOO:BAR', environment variables must be provided as 'key=value'"))
			})
		})

		Context("when given empty input", func() {
			BeforeEach(func() {
				input = []string{}
			})

			It("should not fail and produce empty output", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(output).To(BeEmpty())
			})
		})

		Context("when given a single env var", func() {
			BeforeEach(func() {
				input = []string{"FOO=bar"}
			})

			It("should not fail and produce expected output", func() {
				expected := []v1.EnvVar{
					{
						Name:  "FOO",
						Value: "bar",
					},
				}
				Expect(err).NotTo(HaveOccurred())
				Expect(output).To(Equal(expected))
			})
		})

		Context("when given multiple env vars", func() {
			BeforeEach(func() {
				input = []string{"FOO=bar", "BAZ=foo"}
			})

			It("should not fail and produce expected output", func() {
				expected := []v1.EnvVar{
					{
						Name:  "FOO",
						Value: "bar",
					},
					{
						Name:  "BAZ",
						Value: "foo",
					},
				}
				Expect(err).NotTo(HaveOccurred())
				Expect(output).To(Equal(expected))
			})
		})
	})

	Describe("ParseEnvVarSource", func() {
		JustBeforeEach(func() {
			output, err = core.ParseEnvVarSource(input)
		})

		Context("when key not specified", func() {
			BeforeEach(func() {
				input = []string{"="}
			})

			It("should fail with a suitable error", func() {
				Expect(err).To(MatchError("unable to parse '=', the key part is missing"))
			})
		})

		Context("when equal sign not specified", func() {
			BeforeEach(func() {
				input = []string{"FOO:BAR"}
			})

			It("should fail with a suitable error", func() {
				Expect(err).To(MatchError("unable to parse 'FOO:BAR', environment variables must be provided as 'key=value'"))
			})
		})

		Context("when bad source type specified", func() {
			BeforeEach(func() {
				input = []string{"FOO=secretRef:bar:baz"}
			})

			It("should fail with a suitable error", func() {
				Expect(err).To(MatchError("unable to parse 'env-from' entry 'FOO=secretRef:bar:baz', the only accepted source types are secretKeyRef and configMapKeyRef"))
			})
		})

		Context("when not enough parameters specified for secretKeyRef", func() {
			BeforeEach(func() {
				input = []string{"FOO=secretKeyRef:bar"}
			})

			It("should fail with a suitable error", func() {
				Expect(err).To(MatchError("unable to parse 'env-from' entry 'FOO=secretKeyRef:bar', it should be provided as secretKeyRef:{secret-name}:{key-to-select}"))
			})
		})

		Context("when not enough parameters specified for configMapKeyRef", func() {
			BeforeEach(func() {
				input = []string{"FOO=configMapKeyRef:bar"}
			})

			It("should fail with a suitable error", func() {
				Expect(err).To(MatchError("unable to parse 'env-from' entry 'FOO=configMapKeyRef:bar', it should be provided as configMapKeyRef:{config-map-name}:{key-to-select}"))
			})
		})

		Context("when given empty input", func() {
			BeforeEach(func() {
				input = []string{}
			})

			It("should not fail and produce empty output", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(output).To(BeEmpty())
			})
		})

		Context("when given env var source from a secret", func() {
			BeforeEach(func() {
				input = []string{"FOO=secretKeyRef:bar:baz"}
			})

			It("should not fail and produce expected output", func() {
				expected := []v1.EnvVar{
					{
						Name: "FOO",
						ValueFrom: &v1.EnvVarSource{
							SecretKeyRef: &v1.SecretKeySelector{
								LocalObjectReference: v1.LocalObjectReference{
									Name: "bar",
								},
								Key: "baz",
							},
						},
					},
				}
				Expect(err).NotTo(HaveOccurred())
				Expect(output).To(Equal(expected))
			})
		})

		Context("when given env var source from a configMap", func() {
			BeforeEach(func() {
				input = []string{"FOO=configMapKeyRef:bar:baz"}
			})

			It("should not fail and produce expected output", func() {
				expected := []v1.EnvVar{
					{
						Name: "FOO",
						ValueFrom: &v1.EnvVarSource{
							ConfigMapKeyRef: &v1.ConfigMapKeySelector{
								LocalObjectReference: v1.LocalObjectReference{
									Name: "bar",
								},
								Key: "baz",
							},
						},
					},
				}
				Expect(err).NotTo(HaveOccurred())
				Expect(output).To(Equal(expected))
			})
		})

		Context("when given a combination of env var sources", func() {
			BeforeEach(func() {
				input = []string{
					"MAP=configMapKeyRef:bar:baz",
					"SEC=secretKeyRef:bar:baz"}
			})

			It("should not fail and produce expected output", func() {
				expected := []v1.EnvVar{
					{
						Name: "MAP",
						ValueFrom: &v1.EnvVarSource{
							ConfigMapKeyRef: &v1.ConfigMapKeySelector{
								LocalObjectReference: v1.LocalObjectReference{
									Name: "bar",
								},
								Key: "baz",
							},
						},
					},
					{
						Name: "SEC",
						ValueFrom: &v1.EnvVarSource{
							SecretKeyRef: &v1.SecretKeySelector{
								LocalObjectReference: v1.LocalObjectReference{
									Name: "bar",
								},
								Key: "baz",
							},
						},
					},
				}
				Expect(err).NotTo(HaveOccurred())
				Expect(output).To(Equal(expected))
			})
		})
	})
})
