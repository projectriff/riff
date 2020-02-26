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

package testing

import (
	"context"
	"testing"

	"github.com/go-logr/logr"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	controllerruntime "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/projectriff/riff/system/pkg/controllers"
	"github.com/projectriff/riff/system/pkg/tracker"
)

// SubTestcase holds a single row of a table test.
type SubTestcase struct {
	// Name is a descriptive name for this test suitable as a first argument to t.Run()
	Name string
	// Focus is true if and only if only this and any other focussed tests are to be executed.
	// If one or more tests are focussed, the overall table test will fail.
	Focus bool
	// Skip is true if and only if this test should be skipped.
	Skip bool

	// inputs

	// Parent is the initial object passed to the sub reconciler
	Parent Factory
	// GivenStashedValues adds these items to the stash passed into the reconciler. Factories are resolved to their object.
	GivenStashedValues map[controllers.StashKey]interface{}
	// WithReactors installs each ReactionFunc into each fake clientset. ReactionFuncs intercept
	// each call to the clientset providing the ability to mutate the resource or inject an error.
	WithReactors []ReactionFunc
	// GivenObjects build the kubernetes objects which are present at the onset of reconciliation
	GivenObjects []Factory

	// side effects

	// ExpectParent is the expected parent as mutated after the sub reconciler, or nil if no modification
	ExpectParent Factory
	// ExpectStashedValues ensures each value is stashed. Values in the stash that are not expected are ignored. Factories are resolved to their object.
	ExpectStashedValues map[controllers.StashKey]interface{}
	// ExpectTracks holds the ordered list of Track calls expected during reconciliation
	ExpectTracks []TrackRequest
	// ExpectEvents holds the ordered list of events recorded during the reconciliation
	ExpectEvents []Event
	// ExpectCreates builds the ordered list of objects expected to be created during reconciliation
	ExpectCreates []Factory
	// ExpectUpdates builds the ordered list of objects expected to be updated during reconciliation
	ExpectUpdates []Factory
	// ExpectDeletes holds the ordered list of objects expected to be deleted during reconciliation
	ExpectDeletes []DeleteRef

	// outputs

	// ShouldErr is true if and only if reconciliation is expected to return an error
	ShouldErr bool
	// ExpectedResult is compared to the result returned from the reconciler if there was no error
	ExpectedResult controllerruntime.Result
	// Verify provides the reconciliation Result and error for custom assertions
	Verify VerifyFunc

	// lifecycle

	// Prepare is called before the reconciler is executed. It is intended to prepare the broader
	// environment before the specific table record is executed. For example, setting mock expectations.
	Prepare func(t *testing.T) error
	// CleanUp is called after the table record is finished and all defined assertions complete.
	// It is indended to clean up any state created in the Prepare step or during the test
	// execution, or to make assertions for mocks.
	CleanUp func(t *testing.T) error
}

// SubTable represents a list of Testcase tests instances.
type SubTable []SubTestcase

// Test executes the test for a table row.
func (tc *SubTestcase) Test(t *testing.T, scheme *runtime.Scheme, factory SubReconcilerFactory) {
	t.Helper()
	if tc.Skip {
		t.SkipNow()
	}

	// Record the given objects
	givenObjects := make([]runtime.Object, 0, len(tc.GivenObjects))
	originalGivenObjects := make([]runtime.Object, 0, len(tc.GivenObjects))
	for _, f := range tc.GivenObjects {
		object := f.CreateObject()
		givenObjects = append(givenObjects, object.DeepCopyObject())
		originalGivenObjects = append(originalGivenObjects, object.DeepCopyObject())
	}

	clientWrapper := newClientWrapperWithScheme(scheme, givenObjects...)
	for i := range tc.WithReactors {
		// in reverse order since we prepend
		reactor := tc.WithReactors[len(tc.WithReactors)-1-i]
		clientWrapper.PrependReactor("*", "*", reactor)
	}
	tracker := createTracker()
	recorder := &eventRecorder{
		events: []Event{},
		scheme: scheme,
	}
	log := TestLogger(t)
	c := factory(t, tc, clientWrapper, tracker, recorder, log)

	if tc.CleanUp != nil {
		defer func() {
			if err := tc.CleanUp(t); err != nil {
				t.Errorf("error during clean up: %s", err)
			}
		}()
	}
	if tc.Prepare != nil {
		if err := tc.Prepare(t); err != nil {
			t.Errorf("error during prepare: %s", err)
		}
	}

	ctx := controllers.WithStash(context.Background())
	for k, v := range tc.GivenStashedValues {
		if f, ok := v.(Factory); ok {
			v = f.CreateObject()
		}
		controllers.StashValue(ctx, k, v)
	}

	parent := tc.Parent.CreateObject()

	// Run the Reconcile we're testing.
	result, err := c.Reconcile(ctx, parent)

	if (err != nil) != tc.ShouldErr {
		t.Errorf("Reconcile() error = %v, ExpectErr %v", err, tc.ShouldErr)
	}
	if err == nil {
		// result is only significant if there wasn't an error
		if diff := cmp.Diff(tc.ExpectedResult, result); diff != "" {
			t.Errorf("Unexpected result (-expected, +actual): %s", diff)
		}
	}

	if tc.Verify != nil {
		tc.Verify(t, result, err)
	}

	expectedParent := tc.Parent.CreateObject()
	if tc.ExpectParent != nil {
		expectedParent = tc.ExpectParent.CreateObject()
	}
	if diff := cmp.Diff(expectedParent, parent, ignoreLastTransitionTime, safeDeployDiff, ignoreTypeMeta, cmpopts.EquateEmpty()); diff != "" {
		t.Errorf("Unexpected parent mutations(-expected, +actual): %s", diff)
	}

	for key, expected := range tc.ExpectStashedValues {
		if f, ok := expected.(Factory); ok {
			expected = f.CreateObject()
		}
		actual := controllers.RetrieveValue(ctx, key)
		if diff := cmp.Diff(expected, actual, ignoreLastTransitionTime, safeDeployDiff, ignoreTypeMeta, cmpopts.EquateEmpty()); diff != "" {
			t.Errorf("Unexpected stash value %q (-expected, +actual): %s", key, diff)
		}
	}

	actualTracks := tracker.getTrackRequests()
	for i, exp := range tc.ExpectTracks {
		if i >= len(actualTracks) {
			t.Errorf("Missing tracking request: %s", exp)
			continue
		}

		if diff := cmp.Diff(exp, actualTracks[i]); diff != "" {
			t.Errorf("Unexpected tracking request(-expected, +actual): %s", diff)
		}
	}
	if actual, exp := len(actualTracks), len(tc.ExpectTracks); actual > exp {
		for _, extra := range actualTracks[exp:] {
			t.Errorf("Extra tracking request: %s", extra)
		}
	}

	actualEvents := recorder.events
	for i, exp := range tc.ExpectEvents {
		if i >= len(actualEvents) {
			t.Errorf("Missing recorded event: %s", exp)
			continue
		}

		if diff := cmp.Diff(exp, actualEvents[i]); diff != "" {
			t.Errorf("Unexpected recorded event(-expected, +actual): %s", diff)
		}
	}
	if actual, exp := len(actualEvents), len(tc.ExpectEvents); actual > exp {
		for _, extra := range actualEvents[exp:] {
			t.Errorf("Extra recorded event: %s", extra)
		}
	}

	compareActions(t, "create", tc.ExpectCreates, clientWrapper.createActions, ignoreLastTransitionTime, safeDeployDiff, ignoreTypeMeta, cmpopts.EquateEmpty())
	compareActions(t, "update", tc.ExpectUpdates, clientWrapper.updateActions, ignoreLastTransitionTime, safeDeployDiff, ignoreTypeMeta, cmpopts.EquateEmpty())

	for i, exp := range tc.ExpectDeletes {
		if i >= len(clientWrapper.deleteActions) {
			t.Errorf("Missing delete: %#v", exp)
			continue
		}
		actual := NewDeleteRef(clientWrapper.deleteActions[i])

		if diff := cmp.Diff(exp, actual); diff != "" {
			t.Errorf("Unexpected delete (-expected, +actual): %s", diff)
		}
	}
	if actual, expected := len(clientWrapper.deleteActions), len(tc.ExpectDeletes); actual > expected {
		for _, extra := range clientWrapper.deleteActions[expected:] {
			t.Errorf("Extra delete: %#v", extra)
		}
	}

	// Validate the given objects are not mutated by reconciliation
	if diff := cmp.Diff(originalGivenObjects, givenObjects, safeDeployDiff, cmpopts.EquateEmpty()); diff != "" {
		t.Errorf("Given objects mutated by test %s (-expected, +actual): %v", tc.Name, diff)
	}
}

// Test executes the whole suite of the table tests.
func (tb SubTable) Test(t *testing.T, scheme *runtime.Scheme, factory SubReconcilerFactory) {
	t.Helper()
	focussed := SubTable{}
	for _, test := range tb {
		if test.Focus {
			focussed = append(focussed, test)
			break
		}
	}
	testsToExecute := tb
	if len(focussed) > 0 {
		testsToExecute = focussed
	}
	for _, test := range testsToExecute {
		t.Run(test.Name, func(t *testing.T) {
			t.Helper()
			test.Test(t, scheme, factory)
		})
	}
	if len(focussed) > 0 {
		t.Errorf("%d tests out of %d are still focussed, so the table test fails", len(focussed), len(tb))
	}
}

// SubReconcilerFactory returns a Reconciler.Interface to perform reconciliation in table test,
// ActionRecorderList/EventList to capture k8s actions/events produced during reconciliation
// and FakeStatsReporter to capture stats.
type SubReconcilerFactory func(t *testing.T, row *SubTestcase, client client.Client, tracker tracker.Tracker, recorder record.EventRecorder, log logr.Logger) controllers.SubReconciler
