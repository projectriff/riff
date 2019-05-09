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

type CommandTable []CommandTableRow

type CommandTableRow struct {
	Name string

	// environment
	Config       *riff.Config
	GivenObjects []runtime.Object
	WithReactors []ReactionFunc

	// inputs
	Args []string

	// side effects
	ExpectCreates           []metav1.Object
	ExpectUpdates           []runtime.Object
	ExpectDeletes           []DeleteRef
	ExpectDeleteCollections []DeleteCollectionRef

	// outputs
	ShouldError bool
	Verify      func(*T, string, error)

	// lifecycle
	Prepare func(*riff.Config) error
	Cleanup func(*riff.Config) error
}

func (ct CommandTable) Run(t *T, cmdFactory func(*riff.Config) *cobra.Command) {
	for _, ctr := range ct {
		t.Run(ctr.Name, func(t *T) {
			c := ctr.Config
			if c == nil {
				c = &riff.Config{}
			}
			client := NewClient(ctr.GivenObjects...)
			c.Client = client

			if ctr.Prepare != nil {
				if err := ctr.Prepare(c); err != nil {
					t.Errorf("error during prepare: %s", err)
				}
			}

			// Validate all objects that implement Validatable
			client.PrependReactor("create", "*", func(action testing.Action) (handled bool, ret runtime.Object, err error) {
				return ValidateCreates(context.Background(), action)
			})
			client.PrependReactor("update", "*", func(action testing.Action) (handled bool, ret runtime.Object, err error) {
				return ValidateUpdates(context.Background(), action)
			})

			for i := range ctr.WithReactors {
				// in reverse order since we prepend
				reactor := ctr.WithReactors[len(ctr.WithReactors)-1-i]
				client.PrependReactor("*", "*", reactor)
			}

			cmd := cmdFactory(c)
			output := &bytes.Buffer{}

			cmd.SetArgs(ctr.Args)
			cmd.SetOutput(output)

			err := cmd.Execute()

			if expected, actual := ctr.ShouldError, err != nil; expected != actual {
				if expected {
					t.Errorf("expected command to error, actual %v", err)
				} else {
					t.Errorf("expected command not to error, actual %q", err)
				}
			}

			actions, err := client.ActionRecorderList.ActionsByVerb()
			if err != nil {
				t.Errorf("Error capturing actions by verb: %q", err)
			}

			// Previous state is used to diff resource expected state for update requests that were missed.
			objPrevState := map[string]runtime.Object{}
			for _, o := range ctr.GivenObjects {
				objPrevState[objKey(o)] = o
			}

			for i, expected := range ctr.ExpectCreates {
				if i >= len(actions.Creates) {
					t.Errorf("Missing create: %#v", expected)
					continue
				}
				actual := actions.Creates[i]
				obj := actual.GetObject()
				objPrevState[objKey(obj)] = obj

				if at, et := reflect.TypeOf(obj).String(), reflect.TypeOf(expected).String(); at != et {
					t.Errorf("Unexpected create expected type %q, actually %q", et, at)
				} else if diff := cmp.Diff(expected, obj, ignoreLastTransitionTime, safeDeployDiff, cmpopts.EquateEmpty()); diff != "" {
					t.Errorf("Unexpected create (-expected, +actual): %s", diff)
				}
			}
			if actual, expected := len(actions.Creates), len(ctr.ExpectCreates); actual > expected {
				for _, extra := range actions.Creates[expected:] {
					t.Errorf("Extra create: %#v", extra)
				}
			}

			for i, expected := range ctr.ExpectUpdates {
				if i >= len(actions.Updates) {
					key := objKey(expected)
					oldObj, ok := objPrevState[key]
					if !ok {
						t.Errorf("Object %s was never created: expected: %#v", key, expected)
						continue
					}
					t.Errorf("Missing update for %s (-expected, +prevState): %s", key,
						cmp.Diff(expected, oldObj, ignoreLastTransitionTime, safeDeployDiff, cmpopts.EquateEmpty()))
					continue
				}

				actual := actions.Updates[i]
				obj := actual.GetObject()

				if actual.GetSubresource() != "" {
					t.Errorf("Update was invalid - it should not include a subresource: %#v", actual)
				}

				// Update the object state.
				objPrevState[objKey(obj)] = obj

				if at, et := reflect.TypeOf(obj).String(), reflect.TypeOf(expected).String(); at != et {
					t.Errorf("Unexpected update expected type %q, actually %q", et, at)
				} else if diff := cmp.Diff(expected, obj, ignoreLastTransitionTime, safeDeployDiff, cmpopts.EquateEmpty()); diff != "" {
					t.Errorf("Unexpected update (-expected, +actual): %s", diff)
				}
			}

			if actual, expected := len(actions.Updates), len(ctr.ExpectUpdates); actual > expected {
				for _, extra := range actions.Updates[expected:] {
					t.Errorf("Extra update: %#v", extra)
				}
			}
			for i, expected := range ctr.ExpectDeletes {
				if i >= len(actions.Deletes) {
					t.Errorf("Missing delete: %#v", expected)
					continue
				}
				actual := NewDeleteRef(actions.Deletes[i])
				if diff := cmp.Diff(expected, actual); diff != "" {
					t.Errorf("Unexpected delete (-expected, +actual): %s", diff)
				}
			}
			if actual, expected := len(actions.Deletes), len(ctr.ExpectDeletes); actual > expected {
				for _, extra := range actions.Deletes[expected:] {
					t.Errorf("Extra delete: %#v", extra)
				}
			}

			for i, expected := range ctr.ExpectDeleteCollections {
				if i >= len(actions.DeleteCollections) {
					t.Errorf("Missing delete-collection: %#v", expected)
					continue
				}
				actual := NewDeleteCollectionRef(actions.DeleteCollections[i])
				if diff := cmp.Diff(expected, actual); diff != "" {
					t.Errorf("Unexpected delete collection (-expected, +actual): %s", diff)
				}
			}
			if actual, expected := len(actions.DeleteCollections), len(ctr.ExpectDeleteCollections); actual > expected {
				for _, extra := range actions.DeleteCollections[expected:] {
					t.Errorf("Extra delete-collection: %#v", extra)
				}
			}

			if ctr.Verify != nil {
				ctr.Verify(t, output.String(), err)
			}

			if ctr.Cleanup != nil {
				if err := ctr.Cleanup(c); err != nil {
					t.Errorf("error during cleanup: %s", err)
				}
			}
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

type DeleteRef struct {
	Group     string
	Resource  string
	Namespace string
	Name      string
}

func NewDeleteRef(action DeleteAction) DeleteRef {
	return DeleteRef{
		Group:     action.GetResource().Group,
		Resource:  action.GetResource().Resource,
		Namespace: action.GetNamespace(),
		Name:      action.GetName(),
	}
}

type DeleteCollectionRef struct {
	Group         string
	Resource      string
	Namespace     string
	LabelSelector string
}

func NewDeleteCollectionRef(action DeleteCollectionAction) DeleteCollectionRef {
	return DeleteCollectionRef{
		Group:         action.GetResource().Group,
		Resource:      action.GetResource().Resource,
		Namespace:     action.GetNamespace(),
		LabelSelector: action.GetListRestrictions().Labels.String(),
	}
}
