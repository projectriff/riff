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
	"strings"
	"testing"

	"github.com/go-logr/logr"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	controllerruntime "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/projectriff/riff/system/pkg/tracker"
)

// Testcase holds a single row of a table test.
type Testcase struct {
	// Name is a descriptive name for this test suitable as a first argument to t.Run()
	Name string
	// Focus is true if and only if only this and any other focussed tests are to be executed.
	// If one or more tests are focussed, the overall table test will fail.
	Focus bool
	// Skip is true if and only if this test should be skipped.
	Skip bool

	// inputs

	// Key identifies the object to be reconciled
	Key types.NamespacedName
	// WithReactors installs each ReactionFunc into each fake clientset. ReactionFuncs intercept
	// each call to the clientset providing the ability to mutate the resource or inject an error.
	WithReactors []ReactionFunc
	// GivenObjects build the kubernetes objects which are present at the onset of reconciliation
	GivenObjects []Factory
	// APIGivenObjects contains objects that are only available via an API reader instead of the normal cache
	APIGivenObjects []Factory

	// side effects

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
	// ExpectStatusUpdates builds the ordered list of objects whose status is updated during reconciliation
	ExpectStatusUpdates []Factory

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

// VerifyFunc is a verification function
type VerifyFunc func(t *testing.T, result controllerruntime.Result, err error)

// Table represents a list of Testcase tests instances.
type Table []Testcase

// Test executes the test for a table row.
func (tc *Testcase) Test(t *testing.T, scheme *runtime.Scheme, factory ReconcilerFactory) {
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
	apiGivenObjects := make([]runtime.Object, 0, len(tc.APIGivenObjects))
	for _, f := range tc.APIGivenObjects {
		apiGivenObjects = append(apiGivenObjects, f.CreateObject())
	}

	clientWrapper := newClientWrapperWithScheme(scheme, givenObjects...)
	for i := range tc.WithReactors {
		// in reverse order since we prepend
		reactor := tc.WithReactors[len(tc.WithReactors)-1-i]
		clientWrapper.PrependReactor("*", "*", reactor)
	}
	apiReader := newClientWrapperWithScheme(scheme, apiGivenObjects...)
	tracker := createTracker()
	recorder := &eventRecorder{
		events: []Event{},
		scheme: scheme,
	}
	log := TestLogger(t)
	c := factory(t, tc, clientWrapper, apiReader, tracker, recorder, log)

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

	// Run the Reconcile we're testing.
	result, err := c.Reconcile(reconcile.Request{
		NamespacedName: tc.Key,
	})

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

	compareActions(t, "status update", tc.ExpectStatusUpdates, clientWrapper.statusUpdateActions, statusSubresourceOnly, ignoreLastTransitionTime, safeDeployDiff, cmpopts.EquateEmpty())

	// Validate the given objects are not mutated by reconciliation
	if diff := cmp.Diff(originalGivenObjects, givenObjects, safeDeployDiff, cmpopts.EquateEmpty()); diff != "" {
		t.Errorf("Given objects mutated by test %s (-expected, +actual): %v", tc.Name, diff)
	}
}

func compareActions(t *testing.T, actionName string, expectedActionFactories []Factory, actualActions []objectAction, diffOptions ...cmp.Option) {
	t.Helper()
	for i, exp := range expectedActionFactories {
		if i >= len(actualActions) {
			t.Errorf("Missing %s: %#v", actionName, exp.CreateObject())
			continue
		}
		actual := actualActions[i].GetObject()

		if diff := cmp.Diff(exp.CreateObject(), actual, diffOptions...); diff != "" {
			t.Errorf("Unexpected %s (-expected, +actual): %s", actionName, diff)
		}
	}
	if actual, expected := len(actualActions), len(expectedActionFactories); actual > expected {
		for _, extra := range actualActions[expected:] {
			t.Errorf("Extra %s: %#v", actionName, extra)
		}
	}
}

var (
	ignoreLastTransitionTime = cmp.FilterPath(func(p cmp.Path) bool {
		return strings.HasSuffix(p.String(), "LastTransitionTime.Inner.Time")
	}, cmp.Ignore())
	ignoreTypeMeta = cmp.FilterPath(func(p cmp.Path) bool {
		path := p.String()
		return strings.HasSuffix(path, "TypeMeta.APIVersion") || strings.HasSuffix(path, "TypeMeta.Kind")
	}, cmp.Ignore())

	statusSubresourceOnly = cmp.FilterPath(func(p cmp.Path) bool {
		q := p.String()
		return q != "" && !strings.HasPrefix(q, "Status")
	}, cmp.Ignore())

	safeDeployDiff = cmpopts.IgnoreUnexported(resource.Quantity{})
)

// Test executes the whole suite of the table tests.
func (tb Table) Test(t *testing.T, scheme *runtime.Scheme, factory ReconcilerFactory) {
	t.Helper()
	focussed := Table{}
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

// ReconcilerFactory returns a Reconciler.Interface to perform reconciliation in table test,
// ActionRecorderList/EventList to capture k8s actions/events produced during reconciliation
// and FakeStatsReporter to capture stats.
type ReconcilerFactory func(t *testing.T, row *Testcase, client client.Client, apiReader client.Reader, tracker tracker.Tracker, recorder record.EventRecorder, log logr.Logger) reconcile.Reconciler

type DeleteRef struct {
	Group     string
	Kind      string
	Namespace string
	Name      string
}

func NewDeleteRef(action DeleteAction) DeleteRef {
	return DeleteRef{
		Group:     action.GetResource().Group,
		Kind:      action.GetResource().Resource,
		Namespace: action.GetNamespace(),
		Name:      action.GetName(),
	}
}
