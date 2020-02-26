/*
Copyright 2019 the original author or authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/projectriff/system/pkg/validation"
)

func TestValidateInMemoryGateway(t *testing.T) {
	for _, c := range []struct {
		name     string
		target   *InMemoryGateway
		expected validation.FieldErrors
	}{{
		name:     "empty",
		target:   &InMemoryGateway{},
		expected: validation.FieldErrors{},
	}} {
		t.Run(c.name, func(t *testing.T) {
			actual := c.target.Validate()
			if diff := cmp.Diff(c.expected, actual); diff != "" {
				t.Errorf("validateInMemoryGateway(%s) (-expected, +actual) = %v", c.name, diff)
			}
		})
	}
}

func TestValidateInMemoryGatewaySpec(t *testing.T) {
	for _, c := range []struct {
		name     string
		target   *InMemoryGatewaySpec
		expected validation.FieldErrors
	}{{
		name:     "empty",
		target:   &InMemoryGatewaySpec{},
		expected: validation.FieldErrors{},
	}} {
		t.Run(c.name, func(t *testing.T) {
			actual := c.target.Validate()
			if diff := cmp.Diff(c.expected, actual); diff != "" {
				t.Errorf("validateInMemoryGatewaySpec(%s) (-expected, +actual) = %v", c.name, diff)
			}
		})
	}
}
