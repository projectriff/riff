/*
 * Copyright 2020 the original author or authors.
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
	streamingv1alpha1 "github.com/projectriff/riff/system/pkg/apis/streaming/v1alpha1"
	streamv1alpha1 "github.com/projectriff/riff/system/pkg/apis/streaming/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestGatewayListOptions(t *testing.T) {
	table := rifftesting.OptionsTable{
		{
			Name: "invalid list",
			Options: &commands.GatewayListOptions{
				ListOptions: rifftesting.InvalidListOptions,
			},
			ExpectFieldErrors: rifftesting.InvalidListOptionsFieldError,
		},
		{
			Name: "valid list",
			Options: &commands.GatewayListOptions{
				ListOptions: rifftesting.ValidListOptions,
			},
			ShouldValidate: true,
		},
	}

	table.Run(t)
}

func TestGatewayListCommand(t *testing.T) {
	gatewayName := "test-gateway"
	gatewayOtherName := "test-other-gateway"
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
No gateways found.
`,
		},
		{
			Name: "lists an item",
			Args: []string{},
			GivenObjects: []runtime.Object{
				&streamv1alpha1.Gateway{
					ObjectMeta: metav1.ObjectMeta{
						Name:      gatewayName,
						Namespace: defaultNamespace,
					},
				},
			},
			ExpectOutput: `
NAME           TYPE      STATUS      AGE
test-gateway   <empty>   <unknown>   <unknown>
`,
		},
		{
			Name: "filters by namespace",
			Args: []string{cli.NamespaceFlagName, otherNamespace},
			GivenObjects: []runtime.Object{
				&streamv1alpha1.Gateway{
					ObjectMeta: metav1.ObjectMeta{
						Name:      gatewayName,
						Namespace: defaultNamespace,
					},
				},
			},
			ExpectOutput: `
No gateways found.
`,
		},
		{
			Name: "all namespace",
			Args: []string{cli.AllNamespacesFlagName},
			GivenObjects: []runtime.Object{
				&streamv1alpha1.Gateway{
					ObjectMeta: metav1.ObjectMeta{
						Name:      gatewayName,
						Namespace: defaultNamespace,
					},
				},
				&streamv1alpha1.Gateway{
					ObjectMeta: metav1.ObjectMeta{
						Name:      gatewayOtherName,
						Namespace: otherNamespace,
					},
				},
			},
			ExpectOutput: `
NAMESPACE         NAME                 TYPE      STATUS      AGE
default           test-gateway         <empty>   <unknown>   <unknown>
other-namespace   test-other-gateway   <empty>   <unknown>   <unknown>
`,
		},
		{
			Name: "table populates all columns",
			Args: []string{},
			GivenObjects: []runtime.Object{
				&streamv1alpha1.Gateway{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "my-gateway",
						Namespace: defaultNamespace,
						Labels: map[string]string{
							streamingv1alpha1.GatewayTypeLabelKey: "test-type",
						},
					},
					Status: streamv1alpha1.GatewayStatus{
						Status: apis.Status{
							Conditions: apis.Conditions{
								{Type: streamv1alpha1.GatewayConditionReady, Status: "True"},
							},
						},
					},
				},
			},
			ExpectOutput: `
NAME         TYPE        STATUS   AGE
my-gateway   test-type   Ready    <unknown>
`,
		},
		{
			Name: "list error",
			Args: []string{},
			WithReactors: []rifftesting.ReactionFunc{
				rifftesting.InduceFailure("list", "gatewaies"),
			},
			ShouldError: true,
		},
	}

	table.Run(t, commands.NewGatewayListCommand)
}
