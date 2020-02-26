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

func TestValidateApplication(t *testing.T) {
	for _, c := range []struct {
		name     string
		target   *Application
		expected validation.FieldErrors
	}{{
		name:     "empty",
		target:   &Application{},
		expected: validation.ErrMissingField("spec"),
	}, {
		name: "valid",
		target: &Application{
			Spec: ApplicationSpec{
				Image: "test-image",
				Source: &Source{
					Git: &Git{
						URL:      "https://example.com/repo.git",
						Revision: "master",
					},
				},
			},
		},
		expected: validation.FieldErrors{},
	}} {
		t.Run(c.name, func(t *testing.T) {
			actual := c.target.Validate()
			if diff := cmp.Diff(c.expected, actual); diff != "" {
				t.Errorf("validateApplication(%s) (-expected, +actual) = %v", c.name, diff)
			}
		})
	}
}

func TestValidateApplicationSpec(t *testing.T) {
	for _, c := range []struct {
		name     string
		target   *ApplicationSpec
		expected validation.FieldErrors
	}{{
		name:     "empty",
		target:   &ApplicationSpec{},
		expected: validation.ErrMissingField(validation.CurrentField),
	}, {
		name: "valid",
		target: &ApplicationSpec{
			Image: "test-image",
			Source: &Source{
				Git: &Git{
					URL:      "https://example.com/repo.git",
					Revision: "master",
				},
			},
		},
		expected: validation.FieldErrors{},
	}, {
		name: "requires image",
		target: &ApplicationSpec{
			Source: &Source{
				Git: &Git{
					URL:      "https://example.com/repo.git",
					Revision: "master",
				},
			},
		},
		expected: validation.ErrMissingField("image"),
	}, {
		name: "does not require source",
		target: &ApplicationSpec{
			Image: "test-image",
		},
		expected: validation.FieldErrors{},
	}, {
		name: "validates source",
		target: &ApplicationSpec{
			Image:  "test-image",
			Source: &Source{},
		},
		expected: validation.ErrMissingField("source"),
	}} {
		t.Run(c.name, func(t *testing.T) {
			actual := c.target.Validate()
			if diff := cmp.Diff(c.expected, actual); diff != "" {
				t.Errorf("validateApplicationSpec(%s) (-expected, +actual) = %v", c.name, diff)
			}
		})
	}
}
