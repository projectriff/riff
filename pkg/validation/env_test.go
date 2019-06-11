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

package validation_test

import (
	"testing"

	"github.com/projectriff/riff/pkg/cli"
	rifftesting "github.com/projectriff/riff/pkg/testing"
	"github.com/projectriff/riff/pkg/validation"
)

func TestEnvVar(t *testing.T) {
	tests := []struct {
		name     string
		expected *cli.FieldError
		value    string
	}{{
		name:     "valid",
		expected: cli.EmptyFieldError,
		value:    "MY_VAR=my-value",
	}, {
		name:     "empty",
		expected: cli.ErrInvalidValue("", rifftesting.TestField),
		value:    "",
	}, {
		name:     "missing name",
		expected: cli.ErrInvalidValue("=my-value", rifftesting.TestField),
		value:    "=my-value",
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			expected := test.expected
			actual := validation.EnvVar(test.value, rifftesting.TestField)
			if diff := rifftesting.DiffFieldErrors(expected, actual); diff != "" {
				t.Errorf("%s() = (-expected, +actual): %s", test.name, diff)
			}
		})
	}
}

func TestEnvVars(t *testing.T) {
	tests := []struct {
		name     string
		expected *cli.FieldError
		values   []string
	}{{
		name:     "valid, empty",
		expected: cli.EmptyFieldError,
		values:   []string{},
	}, {
		name:     "valid, not empty",
		expected: cli.EmptyFieldError,
		values:   []string{"MY_VAR=my-value"},
	}, {
		name:     "invalid",
		expected: cli.ErrInvalidValue("", cli.CurrentField).ViaFieldIndex(rifftesting.TestField, 0),
		values:   []string{""},
	}, {
		name: "multiple invalid",
		expected: cli.EmptyFieldError.Also(
			cli.ErrInvalidValue("", cli.CurrentField).ViaFieldIndex(rifftesting.TestField, 0),
			cli.ErrInvalidValue("", cli.CurrentField).ViaFieldIndex(rifftesting.TestField, 1),
		),
		values: []string{"", ""},
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			expected := test.expected
			actual := validation.EnvVars(test.values, rifftesting.TestField)
			if diff := rifftesting.DiffFieldErrors(expected, actual); diff != "" {
				t.Errorf("%s() = (-expected, +actual): %s", test.name, diff)
			}
		})
	}
}

func TestEnvVarFrom(t *testing.T) {
	tests := []struct {
		name     string
		expected *cli.FieldError
		value    string
	}{{
		name:     "valid configmap",
		expected: cli.EmptyFieldError,
		value:    "MY_VAR=configMapKeyRef:my-configmap:my-key",
	}, {
		name:     "valid secret",
		expected: cli.EmptyFieldError,
		value:    "MY_VAR=secretKeyRef:my-secret:my-key",
	}, {
		name:     "empty",
		expected: cli.ErrInvalidValue("", rifftesting.TestField),
		value:    "",
	}, {
		name:     "missing name",
		expected: cli.ErrInvalidValue("=configMapKeyRef:my-configmap:my-key", rifftesting.TestField),
		value:    "=configMapKeyRef:my-configmap:my-key",
	}, {
		name:     "unknown type",
		expected: cli.ErrInvalidValue("MY_VAR=otherKeyRef:my-other:my-key", rifftesting.TestField),
		value:    "MY_VAR=otherKeyRef:my-other:my-key",
	}, {
		name:     "missing resource",
		expected: cli.ErrInvalidValue("MY_VAR=configMapKeyRef::my-key", rifftesting.TestField),
		value:    "MY_VAR=configMapKeyRef::my-key",
	}, {
		name:     "missing key",
		expected: cli.ErrInvalidValue("MY_VAR=configMapKeyRef:my-configmap", rifftesting.TestField),
		value:    "MY_VAR=configMapKeyRef:my-configmap",
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			expected := test.expected
			actual := validation.EnvVarFrom(test.value, rifftesting.TestField)
			if diff := rifftesting.DiffFieldErrors(expected, actual); diff != "" {
				t.Errorf("%s() = (-expected, +actual): %s", test.name, diff)
			}
		})
	}
}

func TestEnvVarFroms(t *testing.T) {
	tests := []struct {
		name     string
		expected *cli.FieldError
		values   []string
	}{{
		name:     "valid, empty",
		expected: cli.EmptyFieldError,
		values:   []string{},
	}, {
		name:     "valid, not empty",
		expected: cli.EmptyFieldError,
		values:   []string{"MY_VAR=configMapKeyRef:my-configmap:my-key"},
	}, {
		name:     "invalid",
		expected: cli.ErrInvalidValue("", cli.CurrentField).ViaFieldIndex(rifftesting.TestField, 0),
		values:   []string{""},
	}, {
		name: "multiple invalid",
		expected: cli.EmptyFieldError.Also(
			cli.ErrInvalidValue("", cli.CurrentField).ViaFieldIndex(rifftesting.TestField, 0),
			cli.ErrInvalidValue("", cli.CurrentField).ViaFieldIndex(rifftesting.TestField, 1),
		),
		values: []string{"", ""},
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			expected := test.expected
			actual := validation.EnvVarFroms(test.values, rifftesting.TestField)
			if diff := rifftesting.DiffFieldErrors(expected, actual); diff != "" {
				t.Errorf("%s() = (-expected, +actual): %s", test.name, diff)
			}
		})
	}
}
