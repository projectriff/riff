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
	corev1 "k8s.io/api/core/v1"

	"github.com/projectriff/system/pkg/validation"
)

func TestValidateStream(t *testing.T) {
	for _, c := range []struct {
		name     string
		target   *Stream
		expected validation.FieldErrors
	}{{
		name:     "empty",
		target:   &Stream{},
		expected: validation.ErrMissingField("spec"),
	}, {
		name: "valid",
		target: &Stream{
			Spec: StreamSpec{
				Gateway:     corev1.LocalObjectReference{Name: "kafka"},
				ContentType: "application/json",
			},
		},
		expected: validation.FieldErrors{},
	}} {
		t.Run(c.name, func(t *testing.T) {
			actual := c.target.Validate()
			if diff := cmp.Diff(c.expected, actual); diff != "" {
				t.Errorf("validateStream(%s) (-expected, +actual) = %v", c.name, diff)
			}
		})
	}
}

func TestValidateStreamSpec(t *testing.T) {
	for _, c := range []struct {
		name     string
		target   *StreamSpec
		expected validation.FieldErrors
	}{{
		name:     "empty",
		target:   &StreamSpec{},
		expected: validation.ErrMissingField(validation.CurrentField),
	}, {
		name: "valid",
		target: &StreamSpec{
			Gateway:     corev1.LocalObjectReference{Name: "kafka"},
			ContentType: "video/mp4",
		},
		expected: validation.FieldErrors{},
	}, {
		name: "valid without explicit content-type",
		target: &StreamSpec{
			Gateway: corev1.LocalObjectReference{Name: "kafka"},
		},
		expected: validation.FieldErrors{},
	}, {
		name: "requires gateway",
		target: &StreamSpec{
			ContentType: "image/*",
		},
		expected: validation.ErrMissingField("gateway"),
	}} {
		t.Run(c.name, func(t *testing.T) {
			actual := c.target.Validate()
			if diff := cmp.Diff(c.expected, actual); diff != "" {
				t.Errorf("validateStreamSpec(%s) (-expected, +actual) = %v", c.name, diff)
			}
		})
	}
}
