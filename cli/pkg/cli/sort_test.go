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

package cli_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/projectriff/riff/cli/pkg/cli"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type StubResource struct {
	metav1.ObjectMeta
}

func TestSortByNamespaceAndName(t *testing.T) {
	var (
		aa = StubResource{metav1.ObjectMeta{Namespace: "namespace-a", Name: "name-a"}}
		ab = StubResource{metav1.ObjectMeta{Namespace: "namespace-a", Name: "name-b"}}
		ba = StubResource{metav1.ObjectMeta{Namespace: "namespace-b", Name: "name-a"}}
		bb = StubResource{metav1.ObjectMeta{Namespace: "namespace-b", Name: "name-b"}}
	)

	tests := []struct {
		name   string
		items  []StubResource
		sorted []StubResource
	}{{
		name:   "empty",
		items:  []StubResource{},
		sorted: []StubResource{},
	}, {
		name:   "single item",
		items:  []StubResource{aa},
		sorted: []StubResource{aa},
	}, {
		name:   "sort by namespace first",
		items:  []StubResource{ba, ab},
		sorted: []StubResource{ab, ba},
	}, {
		name:   "sort by namespace first, presorted",
		items:  []StubResource{ab, ba},
		sorted: []StubResource{ab, ba},
	}, {
		name:   "sort by name for equivlent namespace",
		items:  []StubResource{ab, aa},
		sorted: []StubResource{aa, ab},
	}, {
		name:   "sort by name for equivlent namespace, presorted",
		items:  []StubResource{aa, ab},
		sorted: []StubResource{aa, ab},
	}, {
		name:   "sort by namespace and name",
		items:  []StubResource{bb, ba, ab, aa},
		sorted: []StubResource{aa, ab, ba, bb},
	}, {
		name:   "sort by namespace and name, presorted",
		items:  []StubResource{aa, ab, ba, bb},
		sorted: []StubResource{aa, ab, ba, bb},
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual := test.items
			expected := test.sorted
			cli.SortByNamespaceAndName(actual)

			if diff := cmp.Diff(expected, actual); diff != "" {
				t.Errorf("Unexpected sorting (-expected, +actual): %s", diff)
			}
		})
	}
}
