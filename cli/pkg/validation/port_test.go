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

	"github.com/google/go-cmp/cmp"
	"github.com/projectriff/cli/pkg/cli"
	rifftesting "github.com/projectriff/cli/pkg/testing"
	"github.com/projectriff/cli/pkg/validation"
)

func TestPortValues(t *testing.T) {
	tests := []struct {
		name     string
		expected cli.FieldErrors
		value    int32
	}{{
		name:     "valid",
		expected: cli.FieldErrors{},
		value:    8888,
	}, {
		name:     "low",
		expected: cli.FieldErrors{},
		value:    1,
	}, {
		name:     "high",
		expected: cli.FieldErrors{},
		value:    65535,
	}, {
		name:     "zero",
		expected: cli.ErrInvalidValue("0", rifftesting.TestField),
		value:    0,
	}, {
		name:     "too low",
		expected: cli.ErrInvalidValue("-1", rifftesting.TestField),
		value:    -1,
	}, {
		name:     "too high",
		expected: cli.ErrInvalidValue("65536", rifftesting.TestField),
		value:    65536,
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			expected := test.expected
			actual := validation.PortNumber(test.value, rifftesting.TestField)
			if diff := cmp.Diff(expected, actual); diff != "" {
				t.Errorf("%s() = (-expected, +actual): %s", test.name, diff)
			}
		})
	}
}
