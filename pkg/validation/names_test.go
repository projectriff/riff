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

func TestK8sName(t *testing.T) {
	tests := []struct {
		name     string
		expected *cli.FieldError
		value    string
	}{{
		name:     "valid",
		expected: cli.EmptyFieldError,
		value:    "my-resource",
	}, {
		name:     "empty",
		expected: cli.ErrInvalidValue("", rifftesting.TestField),
		value:    "",
	}, {
		name:     "invalid",
		expected: cli.ErrInvalidValue("/", rifftesting.TestField),
		value:    "/",
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			expected := test.expected
			actual := validation.K8sName(test.value, rifftesting.TestField)
			if diff := rifftesting.DiffFieldErrors(expected, actual); diff != "" {
				t.Errorf("%s() = (-expected, +actual): %s", test.name, diff)
			}
		})
	}
}

func TestK8sNames(t *testing.T) {
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
		values:   []string{"my-resource"},
	}, {
		name:     "invalid",
		expected: cli.ErrInvalidValue("", cli.CurrentField).ViaFieldIndex(rifftesting.TestField, 0),
		values:   []string{""},
	}, {
		name:     "invalid",
		expected: cli.ErrInvalidValue("/", cli.CurrentField).ViaFieldIndex(rifftesting.TestField, 0),
		values:   []string{"/"},
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
			actual := validation.K8sNames(test.values, rifftesting.TestField)
			if diff := rifftesting.DiffFieldErrors(expected, actual); diff != "" {
				t.Errorf("%s() = (-expected, +actual): %s", test.name, diff)
			}
		})
	}
}
