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

	"github.com/projectriff/riff/system/pkg/validation"
)

func TestValidateAdapter(t *testing.T) {
	for _, c := range []struct {
		name     string
		target   *Adapter
		expected validation.FieldErrors
	}{{
		name:     "empty",
		target:   &Adapter{},
		expected: validation.ErrMissingField("spec"),
	}, {
		name: "valid",
		target: &Adapter{
			Spec: AdapterSpec{
				Build: Build{
					FunctionRef: "my-function",
				},
				Target: AdapterTarget{
					ServiceRef: "my-service",
				},
			},
		},
		expected: validation.FieldErrors{},
	}} {
		t.Run(c.name, func(t *testing.T) {
			actual := c.target.Validate()
			if diff := cmp.Diff(c.expected, actual); diff != "" {
				t.Errorf("validateAdapter(%s) (-expected, +actual) = %v", c.name, diff)
			}
		})
	}
}

func TestValidateAdapterSpec(t *testing.T) {
	for _, c := range []struct {
		name     string
		target   *AdapterSpec
		expected validation.FieldErrors
	}{{
		name:     "empty",
		target:   &AdapterSpec{},
		expected: validation.ErrMissingField(validation.CurrentField),
	}, {
		name: "valid",
		target: &AdapterSpec{
			Build: Build{
				FunctionRef: "my-function",
			},
			Target: AdapterTarget{
				ServiceRef: "my-service",
			},
		},
		expected: validation.FieldErrors{},
	}, {
		name: "empty build",
		target: &AdapterSpec{
			Build: Build{},
			Target: AdapterTarget{
				ServiceRef: "my-service",
			},
		},
		expected: validation.ErrMissingField("build"),
	}, {
		name: "empty target",
		target: &AdapterSpec{
			Build: Build{
				FunctionRef: "my-function",
			},
			Target: AdapterTarget{},
		},
		expected: validation.ErrMissingField("target"),
	}} {
		t.Run(c.name, func(t *testing.T) {
			actual := c.target.Validate()
			if diff := cmp.Diff(c.expected, actual); diff != "" {
				t.Errorf("validateAdapterSpec(%s) (-expected, +actual) = %v", c.name, diff)
			}
		})
	}
}
