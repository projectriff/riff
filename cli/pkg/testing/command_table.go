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
	"os"
	"os/exec"
	"path"
	"reflect"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/projectriff/cli/pkg/cli"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgotesting "k8s.io/client-go/testing"
)

// CommandTable provides a declarative model for testing interactions with Kubernetes clientsets
// via Cobra commands.
//
// Fake clientsets are used to stub calls to the Kubernetes API server. GivenObjects populate a
// local cache for clientsets to respond to get and list operations (update and delete will error
// if the object does not exist and create operations will error if the resource does exist).
//
// ExpectCreates and ExpectUpdates each contain objects that are compared directly to resources
// received by the clientsets. ExpectDeletes and ExpectDeleteCollections contain references to the
// resources impacted by the call since these calls do not receive an object.
//
// Errors can be injected into API calls by reactor functions specified in WithReactors. A
// ReactionFunc is able to intercept each clientset operation to observe or mutate the request or
// response.
//
// ShouldError must correctly reflect whether the command is expected to return an error,
// otherwise the testcase will fail. Custom assertions based on the content of the error object
// and the console output from the command are available with the Verify callback.
//
// Advanced state may be configured before and after each record by the Prepare and CleanUp
// callbacks respectively.
type CommandTable []CommandTableRecord

// CommandTableRecord is a single test case within a CommandTable. All state and assertions are
// defined within the record.
type CommandTableRecord struct {

	// Name is used to identify the record in the test results. A sub-test is created for each
	// record with this name.
	Name string
	// Skip suppresses the execution of this test record.
	Skip bool
	// Focus executes this record skipping all unfocused records. The containing test will fail to
	// prevent accidental check-in.
	Focus bool

	// environment

	// Config is passed into the command factory. Mosts tests should not need to set this field.
	// If not specified, a default Config is created with a FakeClient. The Config's client will
	// always be replaced with a FakeClient configured with the given objects and reactors to
	// intercept all calls to the fake clientsets for comparison with the expected operations.
	Config *cli.Config
	// Runtimes are an optional array of runtime names to enabled on the config. If not set, the
	// default runtime set is used.
	Runtimes *[]string

	// GivenObjects represents resources that would already exist within Kubernetes. These
	// resources are passed directly to the fake clientsets.
	GivenObjects []runtime.Object
	// WithReactors installs each ReactionFunc into each fake clientset. ReactionFuncs intercept
	// each call to the clientset providing the ability to mutate the resource or inject an error.
	WithReactors []ReactionFunc
	// ExecHelper is a test case that will intercept exec calls receiving their arguments and
	// environment. The helper is able to control stdio and the exit code of the process. Test
	// cases that need to orchestrate multiple exec calls within a single test should instead use
	// a mock.
	//
	// The value of ExecHelper must map to a test function in the same package taking the form
	// `fmt.Sprintf("TestHelperProcess_%s", ExecHelper)``. The test function should distinguish
	// between test exec invocations and vanilla test calls by the `GO_WANT_HELPER_PROCESS` env.
	//
	// ```
	// func TestHelperProcess_Example(t *testing.T) {
	//     if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
	//         return
	//     }
	//     // insert custom behavior
	//     os.Exit(0)
	// }
	// ```
	ExecHelper string

	// inputs

	// Args are passed directly to cobra before executing the command. This is the primary
	// interface to control the behavior of the cli.
	Args []string
	// Stdin injects stub data to be read via os.Stdin for the command. Tests using stdin are
	// forced to be sequential.
	Stdin []byte

	// side effects

	// ExpectCreates asserts each resource with the resources passed to the Create method of the
	// fake clientsets in order.
	ExpectCreates []runtime.Object
	// ExpectUpdates asserts each resource with the resources passed to the Update method of the
	// fake clientsets in order.
	ExpectUpdates []runtime.Object
	// ExpectDeletes assert references to the Delete method of the fake clientsets in order.
	// Unlike Create and Update, Delete does not receive a full resource, so a reference is used
	// instead. The Group will be blank for 'core' resources. The Resource is not a Kind, but
	// plural lowercase name of the resource.
	ExpectDeletes []DeleteRef
	// ExpectDeleteCollections asserts references to the DeleteCollection method of the fake
	// clientsets in order. DeleteCollections behaves similarly to Deletes. Unlike Delete,
	// DeleteCollection does not contain a resource Name, but may contain a LabelSelector.
	ExpectDeleteCollections []DeleteCollectionRef

	// outputs

	// ShouldError indicates if the table record command execution should return an error. The
	// test will fail if this value does not reflect the returned error.
	ShouldError bool
	// ExpectOutput performs a direct comparison of this content with the command's output showing
	// a diff of any changes. The comparison is ignored for empty strings and ignores a leading
	// new line.
	ExpectOutput string
	// Verify provides the command output and error for custom assertions.
	Verify func(t *testing.T, output string, err error)

	// lifecycle

	// Prepare is called before the command is executed. It is intended to prepare that broader
	// environment before the specific table record is executed. For example, chaning the working
	// directory or setting mock expectations.
	Prepare func(t *testing.T, ctx context.Context, config *cli.Config) (context.Context, error)
	// CleanUp is called after the table record is finished and all defined assertions complete.
	// It is indended to clean up any state created in the Prepare step or during the test
	// execution, or to make assertions for mocks.
	CleanUp func(t *testing.T, ctx context.Context, config *cli.Config) error
}

// Run each record for the table. Tables with a focused record will run only the focused records
// and then fail, to prevent accidental check-in.
func (ct CommandTable) Run(t *testing.T, cmdFactory func(context.Context, *cli.Config) *cobra.Command) {
	focusedTable := CommandTable{}
	for _, ctr := range ct {
		if ctr.Focus == true && ctr.Skip != true {
			focusedTable = append(focusedTable, ctr)
		}
	}
	if len(focusedTable) != 0 {
		for _, ctr := range focusedTable {
			ctr.Run(t, cmdFactory)
		}
		t.Errorf("test run focused on %d record(s), skipped %d record(s)", len(focusedTable), len(ct)-len(focusedTable))
		return
	}

	for _, ctr := range ct {
		ctr.Run(t, cmdFactory)
	}
}

// Run a single table record for the command. It is not common to run a record outside of a table.
func (ctr CommandTableRecord) Run(t *testing.T, cmdFactory func(context.Context, *cli.Config) *cobra.Command) {
	t.Run(ctr.Name, func(t *testing.T) {
		if ctr.Skip {
			t.SkipNow()
		}

		ctx := context.Background()
		c := ctr.Config
		if c == nil {
			c = cli.NewDefaultConfig()
		}
		if ctr.Runtimes != nil {
			c.Runtimes = map[string]bool{}
			for _, runtime := range *ctr.Runtimes {
				c.Runtimes[runtime] = true
			}
		}
		client := NewClient(ctr.GivenObjects...)
		c.Client = client
		if ctr.ExecHelper != "" {
			c.Exec = fakeExecCommand(ctr.ExecHelper)
		}

		if ctr.CleanUp != nil {
			defer func() {
				if err := ctr.CleanUp(t, ctx, c); err != nil {
					t.Errorf("error during clean up: %s", err)
				}
			}()
		}
		if ctr.Prepare != nil {
			var err error
			if ctx, err = ctr.Prepare(t, ctx, c); err != nil {
				t.Errorf("error during prepare: %s", err)
			}
		}

		// Validate all objects that implement Validatable
		client.PrependReactor("create", "*", func(action clientgotesting.Action) (handled bool, ret runtime.Object, err error) {
			return ValidateCreates(context.Background(), action)
		})
		client.PrependReactor("update", "*", func(action clientgotesting.Action) (handled bool, ret runtime.Object, err error) {
			return ValidateUpdates(context.Background(), action)
		})

		for i := range ctr.WithReactors {
			// in reverse order since we prepend
			reactor := ctr.WithReactors[len(ctr.WithReactors)-1-i]
			client.PrependReactor("*", "*", reactor)
		}

		cmd := cmdFactory(ctx, c)
		cmd.SilenceErrors = true
		cmd.SilenceUsage = true
		cmd.SetArgs(ctr.Args)

		c.Stdin = bytes.NewBuffer(ctr.Stdin)
		output := &bytes.Buffer{}
		cmd.SetOutput(output)
		c.Stdout = output
		c.Stderr = output

		cmdErr := cmd.Execute()

		if expected, actual := ctr.ShouldError, cmdErr != nil; expected != actual {
			if expected {
				t.Errorf("expected command to error, actual %v", cmdErr)
			} else {
				t.Errorf("expected command not to error, actual %q", cmdErr)
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

			applyDefaults(expected)
			applyDefaults(obj)

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

			applyDefaults(expected)
			applyDefaults(obj)

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

		outputString := output.String()
		if ctr.ExpectOutput != "" {
			if diff := cmp.Diff(strings.TrimPrefix(ctr.ExpectOutput, "\n"), outputString); diff != "" {
				t.Errorf("Unexpected output (-expected, +actual): %s", diff)
			}
		}

		if ctr.Verify != nil {
			ctr.Verify(t, outputString, cmdErr)
		}
	})
}

func objKey(o runtime.Object) string {
	on := o.(metav1.ObjectMetaAccessor)
	// namespace + name is not unique, and the tests don't populate k8s kind
	// information, so use GoLang's type name as part of the key.
	return path.Join(reflect.TypeOf(o).String(), on.GetObjectMeta().GetNamespace(), on.GetObjectMeta().GetName())
}

var (
	ignoreLastTransitionTime = cmp.FilterPath(func(p cmp.Path) bool {
		return strings.HasSuffix(p.String(), "LastTransitionTime.Inner.Time")
	}, cmp.Ignore())

	safeDeployDiff = cmpopts.IgnoreUnexported(resource.Quantity{})
)

func applyDefaults(o runtime.Object) {
	if d, ok := o.(defaultable); ok {
		d.Default()
	}
}

type DeleteRef struct {
	Group     string
	Resource  string
	Namespace string
	Name      string
}

func NewDeleteRef(action clientgotesting.DeleteAction) DeleteRef {
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

func NewDeleteCollectionRef(action clientgotesting.DeleteCollectionAction) DeleteCollectionRef {
	return DeleteCollectionRef{
		Group:         action.GetResource().Group,
		Resource:      action.GetResource().Resource,
		Namespace:     action.GetNamespace(),
		LabelSelector: action.GetListRestrictions().Labels.String(),
	}
}

func fakeExecCommand(helper string) func(context.Context, string, ...string) *exec.Cmd {
	// pattern derived from https://npf.io/2015/06/testing-exec-command/
	return func(ctx context.Context, command string, args ...string) *exec.Cmd {
		cs := []string{"-test.run=TestHelperProcess_" + helper, "--", command}
		cs = append(cs, args...)
		cmd := exec.CommandContext(ctx, os.Args[0], cs...)
		cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1"}
		return cmd
	}
}

type defaultable interface {
	Default()
}
