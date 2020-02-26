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
	"context"
	"testing"

	"github.com/projectriff/cli/pkg/cli"
	"github.com/projectriff/cli/pkg/knative/commands"
	rifftesting "github.com/projectriff/cli/pkg/testing"
	"github.com/projectriff/system/pkg/apis"
	knativev1alpha1 "github.com/projectriff/system/pkg/apis/knative/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestAdapterListOptions(t *testing.T) {
	table := rifftesting.OptionsTable{
		{
			Name: "invalid list",
			Options: &commands.AdapterListOptions{
				ListOptions: rifftesting.InvalidListOptions,
			},
			ExpectFieldErrors: rifftesting.InvalidListOptionsFieldError,
		},
		{
			Name: "valid list",
			Options: &commands.AdapterListOptions{
				ListOptions: rifftesting.ValidListOptions,
			},
			ShouldValidate: true,
		},
	}

	table.Run(t)
}

func TestAdapterListCommand(t *testing.T) {
	adapterName := "test-adapter"
	adapterOtherName := "test-other-adapter"
	defaultNamespace := "default"
	otherNamespace := "other-namespace"
	applicationName := "my-app"
	functionName := "my-func"
	containerName := "my-container"
	configurationName := "my-configuration"
	serviceName := "my-service"

	table := rifftesting.CommandTable{
		{
			Name: "invalid args",
			Args: []string{},
			Prepare: func(t *testing.T, ctx context.Context, c *cli.Config) (context.Context, error) {
				// disable default namespace
				c.Client.(*rifftesting.FakeClient).Namespace = ""
				return ctx, nil
			},
			ShouldError: true,
		},
		{
			Name: "empty",
			Args: []string{},
			ExpectOutput: `
No adapters found.
`,
		},
		{
			Name: "lists an item",
			Args: []string{},
			GivenObjects: []runtime.Object{
				&knativev1alpha1.Adapter{
					ObjectMeta: metav1.ObjectMeta{
						Name:      adapterName,
						Namespace: defaultNamespace,
					},
				},
			},
			ExpectOutput: `
NAME           BUILD TYPE   BUILD REF   TARGET TYPE   TARGET REF   STATUS      AGE
test-adapter   <unknown>    <unknown>   <unknown>     <unknown>    <unknown>   <unknown>
`,
		},
		{
			Name: "filters by namespace",
			Args: []string{cli.NamespaceFlagName, otherNamespace},
			GivenObjects: []runtime.Object{
				&knativev1alpha1.Adapter{
					ObjectMeta: metav1.ObjectMeta{
						Name:      adapterName,
						Namespace: defaultNamespace,
					},
				},
			},
			ExpectOutput: `
No adapters found.
`,
		},
		{
			Name: "all namespace",
			Args: []string{cli.AllNamespacesFlagName},
			GivenObjects: []runtime.Object{
				&knativev1alpha1.Adapter{
					ObjectMeta: metav1.ObjectMeta{
						Name:      adapterName,
						Namespace: defaultNamespace,
					},
				},
				&knativev1alpha1.Adapter{
					ObjectMeta: metav1.ObjectMeta{
						Name:      adapterOtherName,
						Namespace: otherNamespace,
					},
				},
			},
			ExpectOutput: `
NAMESPACE         NAME                 BUILD TYPE   BUILD REF   TARGET TYPE   TARGET REF   STATUS      AGE
default           test-adapter         <unknown>    <unknown>   <unknown>     <unknown>    <unknown>   <unknown>
other-namespace   test-other-adapter   <unknown>    <unknown>   <unknown>     <unknown>    <unknown>   <unknown>
`,
		},
		{
			Name: "table populates all columns",
			Args: []string{},
			GivenObjects: []runtime.Object{
				&knativev1alpha1.Adapter{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "app",
						Namespace: defaultNamespace,
					},
					Spec: knativev1alpha1.AdapterSpec{
						Build:  knativev1alpha1.Build{ApplicationRef: applicationName},
						Target: knativev1alpha1.AdapterTarget{ServiceRef: serviceName},
					},
					Status: knativev1alpha1.AdapterStatus{
						Status: apis.Status{
							Conditions: apis.Conditions{
								{Type: knativev1alpha1.AdapterConditionReady, Status: "True"},
							},
						},
					},
				},
				&knativev1alpha1.Adapter{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "func",
						Namespace: defaultNamespace,
					},
					Spec: knativev1alpha1.AdapterSpec{
						Build:  knativev1alpha1.Build{FunctionRef: functionName},
						Target: knativev1alpha1.AdapterTarget{ServiceRef: serviceName},
					},
					Status: knativev1alpha1.AdapterStatus{
						Status: apis.Status{
							Conditions: apis.Conditions{
								{Type: knativev1alpha1.AdapterConditionReady, Status: "True"},
							},
						},
					},
				},
				&knativev1alpha1.Adapter{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "container",
						Namespace: defaultNamespace,
					},
					Spec: knativev1alpha1.AdapterSpec{
						Build:  knativev1alpha1.Build{ContainerRef: containerName},
						Target: knativev1alpha1.AdapterTarget{ConfigurationRef: configurationName},
					},
					Status: knativev1alpha1.AdapterStatus{
						Status: apis.Status{
							Conditions: apis.Conditions{
								{Type: knativev1alpha1.AdapterConditionReady, Status: "True"},
							},
						},
					},
				},
			},
			ExpectOutput: `
NAME        BUILD TYPE    BUILD REF      TARGET TYPE     TARGET REF         STATUS   AGE
app         application   my-app         service         my-service         Ready    <unknown>
container   container     my-container   configuration   my-configuration   Ready    <unknown>
func        function      my-func        service         my-service         Ready    <unknown>
`,
		},
		{
			Name: "list error",
			Args: []string{},
			WithReactors: []rifftesting.ReactionFunc{
				rifftesting.InduceFailure("list", "adapters"),
			},
			ShouldError: true,
		},
	}

	table.Run(t, commands.NewAdapterListCommand)
}
