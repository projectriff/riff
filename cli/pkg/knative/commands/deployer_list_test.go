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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestDeployerListOptions(t *testing.T) {
	table := rifftesting.OptionsTable{
		{
			Name: "invalid list",
			Options: &commands.DeployerListOptions{
				ListOptions: rifftesting.InvalidListOptions,
			},
			ExpectFieldErrors: rifftesting.InvalidListOptionsFieldError,
		},
		{
			Name: "valid list",
			Options: &commands.DeployerListOptions{
				ListOptions: rifftesting.ValidListOptions,
			},
			ShouldValidate: true,
		},
	}

	table.Run(t)
}

func TestDeployerListCommand(t *testing.T) {
	deployerName := "test-deployer"
	deployerOtherName := "test-other-deployer"
	defaultNamespace := "default"
	otherNamespace := "other-namespace"

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
No deployers found.
`,
		},
		{
			Name: "lists an item",
			Args: []string{},
			GivenObjects: []runtime.Object{
				&knativev1alpha1.Deployer{
					ObjectMeta: metav1.ObjectMeta{
						Name:      deployerName,
						Namespace: defaultNamespace,
					},
				},
			},
			ExpectOutput: `
NAME            TYPE        REF         URL       STATUS      AGE
test-deployer   <unknown>   <unknown>   <empty>   <unknown>   <unknown>
`,
		},
		{
			Name: "filters by namespace",
			Args: []string{cli.NamespaceFlagName, otherNamespace},
			GivenObjects: []runtime.Object{
				&knativev1alpha1.Deployer{
					ObjectMeta: metav1.ObjectMeta{
						Name:      deployerName,
						Namespace: defaultNamespace,
					},
				},
			},
			ExpectOutput: `
No deployers found.
`,
		},
		{
			Name: "all namespace",
			Args: []string{cli.AllNamespacesFlagName},
			GivenObjects: []runtime.Object{
				&knativev1alpha1.Deployer{
					ObjectMeta: metav1.ObjectMeta{
						Name:      deployerName,
						Namespace: defaultNamespace,
					},
				},
				&knativev1alpha1.Deployer{
					ObjectMeta: metav1.ObjectMeta{
						Name:      deployerOtherName,
						Namespace: otherNamespace,
					},
				},
			},
			ExpectOutput: `
NAMESPACE         NAME                  TYPE        REF         URL       STATUS      AGE
default           test-deployer         <unknown>   <unknown>   <empty>   <unknown>   <unknown>
other-namespace   test-other-deployer   <unknown>   <unknown>   <empty>   <unknown>   <unknown>
`,
		},
		{
			Name: "table populates all columns",
			Args: []string{},
			GivenObjects: []runtime.Object{
				&knativev1alpha1.Deployer{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "img",
						Namespace: defaultNamespace,
					},
					Spec: knativev1alpha1.DeployerSpec{
						Template: &corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{Image: "projectriff/upper"},
								},
							},
						},
					},
					Status: knativev1alpha1.DeployerStatus{
						Status: apis.Status{
							Conditions: apis.Conditions{
								{Type: knativev1alpha1.DeployerConditionReady, Status: "True"},
							},
						},
						Address: &apis.Addressable{
							URL: "img.default.svc.cluster.local",
						},
						URL: "img.default.example.com",
					},
				},
				&knativev1alpha1.Deployer{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "app",
						Namespace: defaultNamespace,
					},
					Spec: knativev1alpha1.DeployerSpec{
						Build:         &knativev1alpha1.Build{ApplicationRef: "petclinic"},
						IngressPolicy: knativev1alpha1.IngressPolicyExternal,
					},
					Status: knativev1alpha1.DeployerStatus{
						Status: apis.Status{
							Conditions: apis.Conditions{
								{Type: knativev1alpha1.DeployerConditionReady, Status: "True"},
							},
						},
						Address: &apis.Addressable{
							URL: "app.default.svc.cluster.local",
						},
						URL: "app.default.example.com",
					},
				},
				&knativev1alpha1.Deployer{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "func",
						Namespace: defaultNamespace,
					},
					Spec: knativev1alpha1.DeployerSpec{
						Build:         &knativev1alpha1.Build{FunctionRef: "square"},
						IngressPolicy: knativev1alpha1.IngressPolicyExternal,
					},
					Status: knativev1alpha1.DeployerStatus{
						Status: apis.Status{
							Conditions: apis.Conditions{
								{Type: knativev1alpha1.DeployerConditionReady, Status: "True"},
							},
						},
						Address: &apis.Addressable{
							URL: "func.default.svc.cluster.local",
						},
						URL: "func.default.example.com",
					},
				},
				&knativev1alpha1.Deployer{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "container",
						Namespace: defaultNamespace,
					},
					Spec: knativev1alpha1.DeployerSpec{
						Build:         &knativev1alpha1.Build{ContainerRef: "busybox"},
						IngressPolicy: knativev1alpha1.IngressPolicyClusterLocal,
					},
					Status: knativev1alpha1.DeployerStatus{
						Status: apis.Status{
							Conditions: apis.Conditions{
								{Type: knativev1alpha1.DeployerConditionReady, Status: "True"},
							},
						},
						Address: &apis.Addressable{
							URL: "container.default.svc.cluster.local",
						},
						URL: "container.default.example.com",
					},
				},
			},
			ExpectOutput: `
NAME        TYPE          REF                 URL                                   STATUS   AGE
app         application   petclinic           app.default.example.com               Ready    <unknown>
container   container     busybox             container.default.svc.cluster.local   Ready    <unknown>
func        function      square              func.default.example.com              Ready    <unknown>
img         image         projectriff/upper   img.default.example.com               Ready    <unknown>
`,
		},
		{
			Name: "list error",
			Args: []string{},
			WithReactors: []rifftesting.ReactionFunc{
				rifftesting.InduceFailure("list", "deployers"),
			},
			ShouldError: true,
		},
		{
			Name: "cluster local ingress policy with missing address",
			Args: []string{},
			GivenObjects: []runtime.Object{
				&knativev1alpha1.Deployer{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "img",
						Namespace: defaultNamespace,
					},
					Spec: knativev1alpha1.DeployerSpec{
						Template: &corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{Image: "projectriff/upper"},
								},
							},
						},
						IngressPolicy: knativev1alpha1.IngressPolicyClusterLocal,
					},
					Status: knativev1alpha1.DeployerStatus{
						Status: apis.Status{
							Conditions: apis.Conditions{
								{Type: knativev1alpha1.DeployerConditionReady, Status: "True"},
							},
						},
						Address: nil,
						URL:     "img.default.example.com",
					},
				},
			},
			ExpectOutput: `
NAME   TYPE    REF                 URL       STATUS   AGE
img    image   projectriff/upper   <empty>   Ready    <unknown>
`,
		},
	}

	table.Run(t, commands.NewDeployerListCommand)
}
