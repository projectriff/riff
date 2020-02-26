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
)

func TestStreamDefault(t *testing.T) {
	tests := []struct {
		name string
		in   *Stream
		want *Stream
	}{{
		name: "empty",
		in:   &Stream{},
		want: &Stream{
			Spec: StreamSpec{
				ContentType: "application/octet-stream",
			},
		},
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := test.in
			got.Default()
			if diff := cmp.Diff(test.want, got); diff != "" {
				t.Errorf("Default (-want, +got) = %v", diff)
			}
		})
	}
}

func TestStreamSpecDefault(t *testing.T) {
	tests := []struct {
		name string
		in   *StreamSpec
		want *StreamSpec
	}{{
		name: "content type is defaulted",
		in:   &StreamSpec{},
		want: &StreamSpec{
			ContentType: "application/octet-stream",
		},
	}, {
		name: "content type is not overwritten",
		in: &StreamSpec{
			ContentType: "application/x-doom",
		},
		want: &StreamSpec{
			ContentType: "application/x-doom",
		},
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := test.in
			got.Default()
			if diff := cmp.Diff(test.want, got); diff != "" {
				t.Errorf("Default (-want, +got) = %v", diff)
			}
		})
	}
}
