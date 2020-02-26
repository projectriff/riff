/*
 * Copyright 2019 The original author or authors
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

package parsers_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/projectriff/riff/cli/pkg/parsers"
	corev1 "k8s.io/api/core/v1"
)

func TestEnvVar(t *testing.T) {
	tests := []struct {
		name     string
		expected corev1.EnvVar
		value    string
	}{{
		name:  "valid",
		value: "MY_VAR=my-value",
		expected: corev1.EnvVar{
			Name:  "MY_VAR",
			Value: "my-value",
		},
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			expected := test.expected
			actual := parsers.EnvVar(test.value)
			if diff := cmp.Diff(expected, actual); diff != "" {
				t.Errorf("%s() = (-expected, +actual): %s", test.name, diff)
			}
		})
	}
}

func TestEnvVarFrom(t *testing.T) {
	tests := []struct {
		name     string
		expected corev1.EnvVar
		value    string
	}{{
		name:  "configmap",
		value: "MY_VAR=configMapKeyRef:my-configmap:my-key",
		expected: corev1.EnvVar{
			Name: "MY_VAR",
			ValueFrom: &corev1.EnvVarSource{
				ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: "my-configmap",
					},
					Key: "my-key",
				},
			},
		},
	}, {
		name:  "secret",
		value: "MY_VAR=secretKeyRef:my-secret:my-key",
		expected: corev1.EnvVar{
			Name: "MY_VAR",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: "my-secret",
					},
					Key: "my-key",
				},
			},
		},
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			expected := test.expected
			actual := parsers.EnvVarFrom(test.value)
			if diff := cmp.Diff(expected, actual); diff != "" {
				t.Errorf("%s() = (-expected, +actual): %s", test.name, diff)
			}
		})
	}
}
