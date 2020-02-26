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

func TestPulsarGatewayDeleteOptions(t *testing.T) {
	table := rifftesting.OptionsTable{
		{
			Name: "invalid delete",
			Options: &commands.PulsarGatewayDeleteOptions{
				DeleteOptions: rifftesting.InvalidDeleteOptions,
			},
			ExpectFieldErrors: rifftesting.InvalidDeleteOptionsFieldError,
		},
		{
			Name: "valid delete",
			Options: &commands.PulsarGatewayDeleteOptions{
				DeleteOptions: rifftesting.ValidDeleteOptions,
			},
			ShouldValidate: true,
		},
	}

	table.Run(t)
}

func TestPulsarGatewayDeleteCommand(t *testing.T) {
	pulsarGatewayName := "test-pulsar-gateway"
	pulsarGatewayOtherName := "test-other-pulsar-gateway"
	defaultNamespace := "default"

	table := rifftesting.CommandTable{
		{
			Name:        "invalid args",
			Args:        []string{},
			ShouldError: true,
		},
		{
			Name: "delete all pulsar gateways",
			Args: []string{cli.AllFlagName},
			GivenObjects: []runtime.Object{
				&streamv1alpha1.PulsarGateway{
					ObjectMeta: metav1.ObjectMeta{
						Name:      pulsarGatewayName,
						Namespace: defaultNamespace,
					},
				},
			},
			ExpectDeleteCollections: []rifftesting.DeleteCollectionRef{{
				Group:     "streaming.projectriff.io",
				Resource:  "pulsargatewaies",
				Namespace: defaultNamespace,
			}},
			ExpectOutput: `
Deleted pulsar gateways in namespace "default"
`,
		},
		{
			Name: "delete all pulsar gateways error",
			Args: []string{cli.AllFlagName},
			GivenObjects: []runtime.Object{
				&streamv1alpha1.PulsarGateway{
					ObjectMeta: metav1.ObjectMeta{
						Name:      pulsarGatewayName,
						Namespace: defaultNamespace,
					},
				},
			},
			WithReactors: []rifftesting.ReactionFunc{
				rifftesting.InduceFailure("delete-collection", "pulsargatewaies"),
			},
			ExpectDeleteCollections: []rifftesting.DeleteCollectionRef{{
				Group:     "streaming.projectriff.io",
				Resource:  "pulsargatewaies",
				Namespace: defaultNamespace,
			}},
			ShouldError: true,
		},
		{
			Name: "delete pulsar gateways",
			Args: []string{pulsarGatewayName},
			GivenObjects: []runtime.Object{
				&streamv1alpha1.PulsarGateway{
					ObjectMeta: metav1.ObjectMeta{
						Name:      pulsarGatewayName,
						Namespace: defaultNamespace,
					},
				},
			},
			ExpectDeletes: []rifftesting.DeleteRef{{
				Group:     "streaming.projectriff.io",
				Resource:  "pulsargatewaies",
				Namespace: defaultNamespace,
				Name:      pulsarGatewayName,
			}},
			ExpectOutput: `
Deleted pulsar gateway "test-pulsar-gateway"
`,
		},
		{
			Name: "delete pulsar gateway",
			Args: []string{pulsarGatewayName, pulsarGatewayOtherName},
			GivenObjects: []runtime.Object{
				&streamv1alpha1.PulsarGateway{
					ObjectMeta: metav1.ObjectMeta{
						Name:      pulsarGatewayName,
						Namespace: defaultNamespace,
					},
				},
				&streamv1alpha1.PulsarGateway{
					ObjectMeta: metav1.ObjectMeta{
						Name:      pulsarGatewayOtherName,
						Namespace: defaultNamespace,
					},
				},
			},
			ExpectDeletes: []rifftesting.DeleteRef{{
				Group:     "streaming.projectriff.io",
				Resource:  "pulsargatewaies",
				Namespace: defaultNamespace,
				Name:      pulsarGatewayName,
			}, {
				Group:     "streaming.projectriff.io",
				Resource:  "pulsargatewaies",
				Namespace: defaultNamespace,
				Name:      pulsarGatewayOtherName,
			}},
			ExpectOutput: `
Deleted pulsar gateway "test-pulsar-gateway"
Deleted pulsar gateway "test-other-pulsar-gateway"
`,
		},
		{
			Name: "stream does not exist",
			Args: []string{pulsarGatewayName},
			ExpectDeletes: []rifftesting.DeleteRef{{
				Group:     "streaming.projectriff.io",
				Resource:  "pulsargatewaies",
				Namespace: defaultNamespace,
				Name:      pulsarGatewayName,
			}},
			ShouldError: true,
		},
		{
			Name: "delete error",
			Args: []string{pulsarGatewayName},
			GivenObjects: []runtime.Object{
				&streamv1alpha1.PulsarGateway{
					ObjectMeta: metav1.ObjectMeta{
						Name:      pulsarGatewayName,
						Namespace: defaultNamespace,
					},
				},
			},
			WithReactors: []rifftesting.ReactionFunc{
				rifftesting.InduceFailure("delete", "pulsargatewaies"),
			},
			ExpectDeletes: []rifftesting.DeleteRef{{
				Group:     "streaming.projectriff.io",
				Resource:  "pulsargatewaies",
				Namespace: defaultNamespace,
				Name:      pulsarGatewayName,
			}},
			ShouldError: true,
		},
	}

	table.Run(t, commands.NewPulsarGatewayDeleteCommand)
}
