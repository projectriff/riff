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

	"github.com/projectriff/cli/pkg/cli"
	"github.com/projectriff/cli/pkg/streaming/commands"
	rifftesting "github.com/projectriff/cli/pkg/testing"
	streamv1alpha1 "github.com/projectriff/system/pkg/apis/streaming/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestInMemoryGatewayDeleteOptions(t *testing.T) {
	table := rifftesting.OptionsTable{
		{
			Name: "invalid delete",
			Options: &commands.InMemoryGatewayDeleteOptions{
				DeleteOptions: rifftesting.InvalidDeleteOptions,
			},
			ExpectFieldErrors: rifftesting.InvalidDeleteOptionsFieldError,
		},
		{
			Name: "valid delete",
			Options: &commands.InMemoryGatewayDeleteOptions{
				DeleteOptions: rifftesting.ValidDeleteOptions,
			},
			ShouldValidate: true,
		},
	}

	table.Run(t)
}

func TestInMemoryGatewayDeleteCommand(t *testing.T) {
	inmemoryGatewayName := "test-inmemory-gateway"
	inmemoryGatewayOtherName := "test-other-inmemory-gateway"
	defaultNamespace := "default"

	table := rifftesting.CommandTable{
		{
			Name:        "invalid args",
			Args:        []string{},
			ShouldError: true,
		},
		{
			Name: "delete all in-memory gateways",
			Args: []string{cli.AllFlagName},
			GivenObjects: []runtime.Object{
				&streamv1alpha1.InMemoryGateway{
					ObjectMeta: metav1.ObjectMeta{
						Name:      inmemoryGatewayName,
						Namespace: defaultNamespace,
					},
				},
			},
			ExpectDeleteCollections: []rifftesting.DeleteCollectionRef{{
				Group:     "streaming.projectriff.io",
				Resource:  "inmemorygatewaies",
				Namespace: defaultNamespace,
			}},
			ExpectOutput: `
Deleted in-memory gateways in namespace "default"
`,
		},
		{
			Name: "delete all in-memory gateways error",
			Args: []string{cli.AllFlagName},
			GivenObjects: []runtime.Object{
				&streamv1alpha1.InMemoryGateway{
					ObjectMeta: metav1.ObjectMeta{
						Name:      inmemoryGatewayName,
						Namespace: defaultNamespace,
					},
				},
			},
			WithReactors: []rifftesting.ReactionFunc{
				rifftesting.InduceFailure("delete-collection", "inmemorygatewaies"),
			},
			ExpectDeleteCollections: []rifftesting.DeleteCollectionRef{{
				Group:     "streaming.projectriff.io",
				Resource:  "inmemorygatewaies",
				Namespace: defaultNamespace,
			}},
			ShouldError: true,
		},
		{
			Name: "delete in-memory gateways",
			Args: []string{inmemoryGatewayName},
			GivenObjects: []runtime.Object{
				&streamv1alpha1.InMemoryGateway{
					ObjectMeta: metav1.ObjectMeta{
						Name:      inmemoryGatewayName,
						Namespace: defaultNamespace,
					},
				},
			},
			ExpectDeletes: []rifftesting.DeleteRef{{
				Group:     "streaming.projectriff.io",
				Resource:  "inmemorygatewaies",
				Namespace: defaultNamespace,
				Name:      inmemoryGatewayName,
			}},
			ExpectOutput: `
Deleted in-memory gateway "test-inmemory-gateway"
`,
		},
		{
			Name: "delete in-memory gateway",
			Args: []string{inmemoryGatewayName, inmemoryGatewayOtherName},
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
						Namespace: defaultNamespace,
					},
				},
			},
			ExpectDeletes: []rifftesting.DeleteRef{{
				Group:     "streaming.projectriff.io",
				Resource:  "inmemorygatewaies",
				Namespace: defaultNamespace,
				Name:      inmemoryGatewayName,
			}, {
				Group:     "streaming.projectriff.io",
				Resource:  "inmemorygatewaies",
				Namespace: defaultNamespace,
				Name:      inmemoryGatewayOtherName,
			}},
			ExpectOutput: `
Deleted in-memory gateway "test-inmemory-gateway"
Deleted in-memory gateway "test-other-inmemory-gateway"
`,
		},
		{
			Name: "stream does not exist",
			Args: []string{inmemoryGatewayName},
			ExpectDeletes: []rifftesting.DeleteRef{{
				Group:     "streaming.projectriff.io",
				Resource:  "inmemorygatewaies",
				Namespace: defaultNamespace,
				Name:      inmemoryGatewayName,
			}},
			ShouldError: true,
		},
		{
			Name: "delete error",
			Args: []string{inmemoryGatewayName},
			GivenObjects: []runtime.Object{
				&streamv1alpha1.InMemoryGateway{
					ObjectMeta: metav1.ObjectMeta{
						Name:      inmemoryGatewayName,
						Namespace: defaultNamespace,
					},
				},
			},
			WithReactors: []rifftesting.ReactionFunc{
				rifftesting.InduceFailure("delete", "inmemorygatewaies"),
			},
			ExpectDeletes: []rifftesting.DeleteRef{{
				Group:     "streaming.projectriff.io",
				Resource:  "inmemorygatewaies",
				Namespace: defaultNamespace,
				Name:      inmemoryGatewayName,
			}},
			ShouldError: true,
		},
	}

	table.Run(t, commands.NewInMemoryGatewayDeleteCommand)
}
