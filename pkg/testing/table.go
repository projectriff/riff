/*
 * Copyright 2019 the original author or authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      https://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package testing

import (
	"bytes"
	"context"
	"path"
	"reflect"
	"strings"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/knative/pkg/kmeta"
	"github.com/projectriff/riff/pkg/riff"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/testing"
)

type Table []TableRow

type TableRow struct {
	Name                  string
	Args                  []string
	Params                *riff.Params
	Objects               []runtime.Object
	WithReactors          []ReactionFunc
	WantCreates           []metav1.Object
	WantUpdates           []UpdateActionImpl
	WantDeletes           []DeleteActionImpl
	WantDeleteCollections []DeleteCollectionActionImpl
	WantError             bool
	WithOutput            func(*T, string, error)
}

func (tests Table) Run(t *T, cmdFactory func(*riff.Params) *cobra.Command) {
	for _, test := range tests {
		t.Run(test.Name, func(t *T) {
			p := test.Params
			if p == nil {
				p = &riff.Params{}
			}
			client := NewClient(test.Objects...)
			p.Client = client

			// Validate all objects that implement Validatable
			client.PrependReactor("create", "*", func(action testing.Action) (handled bool, ret runtime.Object, err error) {
				return ValidateCreates(context.Background(), action)
			})
			client.PrependReactor("update", "*", func(action testing.Action) (handled bool, ret runtime.Object, err error) {
				return ValidateUpdates(context.Background(), action)
			})

			for i := range test.WithReactors {
				// in reverse order since we prepend
				reactor := test.WithReactors[len(test.WithReactors)-1-i]
				client.PrependReactor("*", "*", reactor)
			}

			cmd := cmdFactory(p)
			output := &bytes.Buffer{}

			cmd.SetArgs(test.Args)
			cmd.SetOutput(output)

			err := cmd.Execute()

			if want, got := test.WantError, err != nil; want != got {
				if want {
					t.Errorf("expected command to error, got %v", got)
				} else {
					t.Errorf("expected command not to error, got %v", got)
				}
			}

			actions, err := client.ActionRecorderList.ActionsByVerb()
			if err != nil {
				t.Errorf("Error capturing actions by verb: %q", err)
			}

			// Previous state is used to diff resource expected state for update requests that were missed.
			objPrevState := map[string]runtime.Object{}
			for _, o := range test.Objects {
				objPrevState[objKey(o)] = o
			}

			for i, want := range test.WantCreates {
				if i >= len(actions.Creates) {
					t.Errorf("Missing create: %#v", want)
					continue
				}
				got := actions.Creates[i]
				obj := got.GetObject()
				objPrevState[objKey(obj)] = obj

				if diff := cmp.Diff(want, obj, ignoreLastTransitionTime, safeDeployDiff, cmpopts.EquateEmpty()); diff != "" {
					t.Errorf("Unexpected create (-want, +got): %s", diff)
				}
			}
			if got, want := len(actions.Creates), len(test.WantCreates); got > want {
				for _, extra := range actions.Creates[want:] {
					t.Errorf("Extra create: %#v", extra)
				}
			}

			for i, want := range test.WantUpdates {
				if i >= len(actions.Updates) {
					wo := want.GetObject()
					key := objKey(wo)
					oldObj, ok := objPrevState[key]
					if !ok {
						t.Errorf("Object %s was never created: want: %#v", key, wo)
						continue
					}
					t.Errorf("Missing update for %s (-want, +prevState): %s", key,
						cmp.Diff(wo, oldObj, ignoreLastTransitionTime, safeDeployDiff, cmpopts.EquateEmpty()))
					continue
				}

				if want.GetSubresource() != "" {
					t.Errorf("Expectation was invalid - it should not include a subresource: %#v", want)
				}

				got := actions.Updates[i].GetObject()

				// Update the object state.
				objPrevState[objKey(got)] = got

				if diff := cmp.Diff(want.GetObject(), got, ignoreLastTransitionTime, safeDeployDiff, cmpopts.EquateEmpty()); diff != "" {
					t.Errorf("Unexpected update (-want, +got): %s", diff)
				}
			}

			if got, want := len(actions.Updates), len(test.WantUpdates); got > want {
				for _, extra := range actions.Updates[want:] {
					t.Errorf("Extra update: %#v", extra)
				}
			}
			for i, want := range test.WantDeletes {
				if i >= len(actions.Deletes) {
					t.Errorf("Missing delete: %#v", want)
					continue
				}
				got := actions.Deletes[i]
				if got.GetName() != want.GetName() || got.GetNamespace() != want.GetNamespace() {
					t.Errorf("Unexpected delete[%d]: %#v", i, got)
				}
				if diff := cmp.Diff(want.GetResource(), got.GetResource()); diff != "" {
					t.Errorf("Unexpected delete (-want, +got): %s", diff)
				}
			}
			if got, want := len(actions.Deletes), len(test.WantDeletes); got > want {
				for _, extra := range actions.Deletes[want:] {
					t.Errorf("Extra delete: %#v", extra)
				}
			}

			for i, want := range test.WantDeleteCollections {
				if i >= len(actions.DeleteCollections) {
					t.Errorf("Missing delete-collection: %#v", want)
					continue
				}
				got := actions.DeleteCollections[i]
				if got, want := got.GetListRestrictions().Labels, want.GetListRestrictions().Labels; (got != nil) != (want != nil) || got.String() != want.String() {
					t.Errorf("Unexpected delete-collection[%d].Labels = %v, wanted %v", i, got, want)
				}
				if got, want := got.GetNamespace(), want.GetNamespace(); got != want {
					t.Errorf("Unexpected delete-collection[%d].Namespace: %#v, wanted %s", i, got, want)
				}
			}
			if got, want := len(actions.DeleteCollections), len(test.WantDeleteCollections); got > want {
				for _, extra := range actions.DeleteCollections[want:] {
					t.Errorf("Extra delete-collection: %#v", extra)
				}
			}

			if test.WithOutput != nil {
				test.WithOutput(t, output.String(), err)
			}
			// TODO assert created, updated and deleted resources
		})
	}
}

func objKey(o runtime.Object) string {
	on := o.(kmeta.Accessor)
	// namespace + name is not unique, and the tests don't populate k8s kind
	// information, so use GoLang's type name as part of the key.
	return path.Join(reflect.TypeOf(o).String(), on.GetNamespace(), on.GetName())
}

var (
	ignoreLastTransitionTime = cmp.FilterPath(func(p cmp.Path) bool {
		return strings.HasSuffix(p.String(), "LastTransitionTime.Inner.Time")
	}, cmp.Ignore())

	safeDeployDiff = cmpopts.IgnoreUnexported(resource.Quantity{})
)
