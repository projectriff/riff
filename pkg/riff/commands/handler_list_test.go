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

package commands_test

import (
	"testing"

	"github.com/knative/pkg/apis"
	duckv1alpha1 "github.com/knative/pkg/apis/duck/v1alpha1"
	"github.com/projectriff/riff/pkg/cli"
	"github.com/projectriff/riff/pkg/riff/commands"
	rifftesting "github.com/projectriff/riff/pkg/testing"
	requestv1alpha1 "github.com/projectriff/system/pkg/apis/request/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestHandlerListOptions(t *testing.T) {
	table := rifftesting.OptionsTable{
		{
			Name: "invalid list",
			Options: &commands.HandlerListOptions{
				ListOptions: rifftesting.InvalidListOptions,
			},
			ExpectFieldError: rifftesting.InvalidListOptionsFieldError,
		},
		{
			Name: "valid list",
			Options: &commands.HandlerListOptions{
				ListOptions: rifftesting.ValidListOptions,
			},
			ShouldValidate: true,
		},
	}

	table.Run(t)
}

func TestHandlerListCommand(t *testing.T) {
	handlerName := "test-handler"
	handlerOtherName := "test-other-handler"
	defaultNamespace := "default"
	otherNamespace := "other-namespace"

	table := rifftesting.CommandTable{
		{
			Name: "invalid args",
			Args: []string{},
			Prepare: func(t *testing.T, c *cli.Config) error {
				// disable default namespace
				c.Client.(*rifftesting.FakeClient).Namespace = ""
				return nil
			},
			ShouldError: true,
		},
		{
			Name: "empty",
			Args: []string{},
			ExpectOutput: `
No handlers found.
`,
		},
		{
			Name: "lists an item",
			Args: []string{},
			GivenObjects: []runtime.Object{
				&requestv1alpha1.Handler{
					ObjectMeta: metav1.ObjectMeta{
						Name:      handlerName,
						Namespace: defaultNamespace,
					},
				},
			},
			ExpectOutput: `
NAME           TYPE        REF         HOST      STATUS      AGE
test-handler   <unknown>   <unknown>   <empty>   <unknown>   <unknown>
`,
		},
		{
			Name: "filters by namespace",
			Args: []string{cli.NamespaceFlagName, otherNamespace},
			GivenObjects: []runtime.Object{
				&requestv1alpha1.Handler{
					ObjectMeta: metav1.ObjectMeta{
						Name:      handlerName,
						Namespace: defaultNamespace,
					},
				},
			},
			ExpectOutput: `
No handlers found.
`,
		},
		{
			Name: "all namespace",
			Args: []string{cli.AllNamespacesFlagName},
			GivenObjects: []runtime.Object{
				&requestv1alpha1.Handler{
					ObjectMeta: metav1.ObjectMeta{
						Name:      handlerName,
						Namespace: defaultNamespace,
					},
				},
				&requestv1alpha1.Handler{
					ObjectMeta: metav1.ObjectMeta{
						Name:      handlerOtherName,
						Namespace: otherNamespace,
					},
				},
			},
			ExpectOutput: `
NAMESPACE         NAME                 TYPE        REF         HOST      STATUS      AGE
default           test-handler         <unknown>   <unknown>   <empty>   <unknown>   <unknown>
other-namespace   test-other-handler   <unknown>   <unknown>   <empty>   <unknown>   <unknown>
`,
		},
		{
			Name: "table populates all columns",
			Args: []string{},
			GivenObjects: []runtime.Object{
				&requestv1alpha1.Handler{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "img",
						Namespace: defaultNamespace,
					},
					Spec: requestv1alpha1.HandlerSpec{
						Template: &corev1.PodSpec{
							Containers: []corev1.Container{
								{Image: "projectriff/upper"},
							},
						},
					},
					Status: requestv1alpha1.HandlerStatus{
						Status: duckv1alpha1.Status{
							Conditions: []duckv1alpha1.Condition{
								{Type: requestv1alpha1.HandlerConditionReady, Status: "True"},
							},
						},
						URL: &apis.URL{
							Host: "image.default.example.com",
						},
					},
				},
				&requestv1alpha1.Handler{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "app",
						Namespace: defaultNamespace,
					},
					Spec: requestv1alpha1.HandlerSpec{
						Build: &requestv1alpha1.Build{ApplicationRef: "petclinic"},
					},
					Status: requestv1alpha1.HandlerStatus{
						Status: duckv1alpha1.Status{
							Conditions: []duckv1alpha1.Condition{
								{Type: requestv1alpha1.HandlerConditionReady, Status: "True"},
							},
						},
						URL: &apis.URL{
							Host: "app.default.example.com",
						},
					},
				},
				&requestv1alpha1.Handler{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "func",
						Namespace: defaultNamespace,
					},
					Spec: requestv1alpha1.HandlerSpec{
						Build: &requestv1alpha1.Build{FunctionRef: "square"},
					},
					Status: requestv1alpha1.HandlerStatus{
						Status: duckv1alpha1.Status{
							Conditions: []duckv1alpha1.Condition{
								{Type: requestv1alpha1.HandlerConditionReady, Status: "True"},
							},
						},
						URL: &apis.URL{
							Host: "func.default.example.com",
						},
					},
				},
			},
			ExpectOutput: `
NAME   TYPE          REF                 HOST                        STATUS   AGE
app    application   petclinic           app.default.example.com     Ready    <unknown>
func   function      square              func.default.example.com    Ready    <unknown>
img    image         projectriff/upper   image.default.example.com   Ready    <unknown>
`,
		},
		{
			Name: "list error",
			Args: []string{},
			WithReactors: []rifftesting.ReactionFunc{
				rifftesting.InduceFailure("list", "handlers"),
			},
			ShouldError: true,
		},
	}

	table.Run(t, commands.NewHandlerListCommand)
}
