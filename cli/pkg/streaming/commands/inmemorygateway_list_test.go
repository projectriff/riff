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

	"github.com/projectriff/riff/cli/pkg/cli"
	"github.com/projectriff/riff/cli/pkg/streaming/commands"
	rifftesting "github.com/projectriff/riff/cli/pkg/testing"
	"github.com/projectriff/riff/system/pkg/apis"
	streamv1alpha1 "github.com/projectriff/riff/system/pkg/apis/streaming/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestInMemoryGatewayListOptions(t *testing.T) {
	table := rifftesting.OptionsTable{
		{
			Name: "invalid list",
			Options: &commands.InMemoryGatewayListOptions{
				ListOptions: rifftesting.InvalidListOptions,
			},
			ExpectFieldErrors: rifftesting.InvalidListOptionsFieldError,
		},
		{
			Name: "valid list",
			Options: &commands.InMemoryGatewayListOptions{
				ListOptions: rifftesting.ValidListOptions,
			},
			ShouldValidate: true,
		},
	}

	table.Run(t)
}

func TestInMemoryGatewayListCommand(t *testing.T) {
	inmemoryGatewayName := "test-inmemory-gateway"
	inmemoryGatewayOtherName := "test-other-inmemory-gateway"
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
No in-memory gateways found.
`,
		},
		{
			Name: "lists an item",
			Args: []string{},
			GivenObjects: []runtime.Object{
				&streamv1alpha1.InMemoryGateway{
					ObjectMeta: metav1.ObjectMeta{
						Name:      inmemoryGatewayName,
						Namespace: defaultNamespace,
					},
				},
			},
			ExpectOutput: `
NAME                    STATUS      AGE
test-inmemory-gateway   <unknown>   <unknown>
`,
		},
		{
			Name: "filters by namespace",
			Args: []string{cli.NamespaceFlagName, otherNamespace},
			GivenObjects: []runtime.Object{
				&streamv1alpha1.InMemoryGateway{
					ObjectMeta: metav1.ObjectMeta{
						Name:      inmemoryGatewayName,
						Namespace: defaultNamespace,
					},
				},
			},
			ExpectOutput: `
No in-memory gateways found.
`,
		},
		{
			Name: "all namespace",
			Args: []string{cli.AllNamespacesFlagName},
			GivenObjects: []runtime.Object{
				&streamv1alpha1.InMemoryGateway{
					ObjectMeta: metav1.ObjectMeta{
						Name:      inmemoryGatewayName,
						Namespace: defaultNamespace,
					},
				},
				&streamv1alpha1.InMemoryGateway{
					ObjectMeta: metav1.ObjectMeta{
						Name:      inmemoryGatewayOtherName,
						Namespace: otherNamespace,
					},
				},
			},
			ExpectOutput: `
NAMESPACE         NAME                          STATUS      AGE
default           test-inmemory-gateway         <unknown>   <unknown>
other-namespace   test-other-inmemory-gateway   <unknown>   <unknown>
`,
		},
		{
			Name: "table populates all columns",
			Args: []string{},
			GivenObjects: []runtime.Object{
				&streamv1alpha1.InMemoryGateway{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "my-inmemory",
						Namespace: defaultNamespace,
					},
					Spec: streamv1alpha1.InMemoryGatewaySpec{},
					Status: streamv1alpha1.InMemoryGatewayStatus{
						Status: apis.Status{
							Conditions: apis.Conditions{
								{Type: streamv1alpha1.InMemoryGatewayConditionReady, Status: "True"},
							},
						},
					},
				},
			},
			ExpectOutput: `
NAME          STATUS   AGE
my-inmemory   Ready    <unknown>
`,
		},
		{
			Name: "list error",
			Args: []string{},
			WithReactors: []rifftesting.ReactionFunc{
				rifftesting.InduceFailure("list", "inmemorygatewaies"),
			},
			ShouldError: true,
		},
	}

	table.Run(t, commands.NewInMemoryGatewayListCommand)
}
