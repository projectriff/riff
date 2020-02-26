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

func TestKafkaGatewayDeleteOptions(t *testing.T) {
	table := rifftesting.OptionsTable{
		{
			Name: "invalid delete",
			Options: &commands.KafkaGatewayDeleteOptions{
				DeleteOptions: rifftesting.InvalidDeleteOptions,
			},
			ExpectFieldErrors: rifftesting.InvalidDeleteOptionsFieldError,
		},
		{
			Name: "valid delete",
			Options: &commands.KafkaGatewayDeleteOptions{
				DeleteOptions: rifftesting.ValidDeleteOptions,
			},
			ShouldValidate: true,
		},
	}

	table.Run(t)
}

func TestKafkaGatewayDeleteCommand(t *testing.T) {
	kafkaGatewayName := "test-kafka-gateway"
	kafkaGatewayOtherName := "test-other-kafka-gateway"
	defaultNamespace := "default"

	table := rifftesting.CommandTable{
		{
			Name:        "invalid args",
			Args:        []string{},
			ShouldError: true,
		},
		{
			Name: "delete all kafka gateways",
			Args: []string{cli.AllFlagName},
			GivenObjects: []runtime.Object{
				&streamv1alpha1.KafkaGateway{
					ObjectMeta: metav1.ObjectMeta{
						Name:      kafkaGatewayName,
						Namespace: defaultNamespace,
					},
				},
			},
			ExpectDeleteCollections: []rifftesting.DeleteCollectionRef{{
				Group:     "streaming.projectriff.io",
				Resource:  "kafkagatewaies",
				Namespace: defaultNamespace,
			}},
			ExpectOutput: `
Deleted kafka gateways in namespace "default"
`,
		},
		{
			Name: "delete all kafka gateways error",
			Args: []string{cli.AllFlagName},
			GivenObjects: []runtime.Object{
				&streamv1alpha1.KafkaGateway{
					ObjectMeta: metav1.ObjectMeta{
						Name:      kafkaGatewayName,
						Namespace: defaultNamespace,
					},
				},
			},
			WithReactors: []rifftesting.ReactionFunc{
				rifftesting.InduceFailure("delete-collection", "kafkagatewaies"),
			},
			ExpectDeleteCollections: []rifftesting.DeleteCollectionRef{{
				Group:     "streaming.projectriff.io",
				Resource:  "kafkagatewaies",
				Namespace: defaultNamespace,
			}},
			ShouldError: true,
		},
		{
			Name: "delete kafka gateways",
			Args: []string{kafkaGatewayName},
			GivenObjects: []runtime.Object{
				&streamv1alpha1.KafkaGateway{
					ObjectMeta: metav1.ObjectMeta{
						Name:      kafkaGatewayName,
						Namespace: defaultNamespace,
					},
				},
			},
			ExpectDeletes: []rifftesting.DeleteRef{{
				Group:     "streaming.projectriff.io",
				Resource:  "kafkagatewaies",
				Namespace: defaultNamespace,
				Name:      kafkaGatewayName,
			}},
			ExpectOutput: `
Deleted kafka gateway "test-kafka-gateway"
`,
		},
		{
			Name: "delete kafka gateway",
			Args: []string{kafkaGatewayName, kafkaGatewayOtherName},
			GivenObjects: []runtime.Object{
				&streamv1alpha1.KafkaGateway{
					ObjectMeta: metav1.ObjectMeta{
						Name:      kafkaGatewayName,
						Namespace: defaultNamespace,
					},
				},
				&streamv1alpha1.KafkaGateway{
					ObjectMeta: metav1.ObjectMeta{
						Name:      kafkaGatewayOtherName,
						Namespace: defaultNamespace,
					},
				},
			},
			ExpectDeletes: []rifftesting.DeleteRef{{
				Group:     "streaming.projectriff.io",
				Resource:  "kafkagatewaies",
				Namespace: defaultNamespace,
				Name:      kafkaGatewayName,
			}, {
				Group:     "streaming.projectriff.io",
				Resource:  "kafkagatewaies",
				Namespace: defaultNamespace,
				Name:      kafkaGatewayOtherName,
			}},
			ExpectOutput: `
Deleted kafka gateway "test-kafka-gateway"
Deleted kafka gateway "test-other-kafka-gateway"
`,
		},
		{
			Name: "stream does not exist",
			Args: []string{kafkaGatewayName},
			ExpectDeletes: []rifftesting.DeleteRef{{
				Group:     "streaming.projectriff.io",
				Resource:  "kafkagatewaies",
				Namespace: defaultNamespace,
				Name:      kafkaGatewayName,
			}},
			ShouldError: true,
		},
		{
			Name: "delete error",
			Args: []string{kafkaGatewayName},
			GivenObjects: []runtime.Object{
				&streamv1alpha1.KafkaGateway{
					ObjectMeta: metav1.ObjectMeta{
						Name:      kafkaGatewayName,
						Namespace: defaultNamespace,
					},
				},
			},
			WithReactors: []rifftesting.ReactionFunc{
				rifftesting.InduceFailure("delete", "kafkagatewaies"),
			},
			ExpectDeletes: []rifftesting.DeleteRef{{
				Group:     "streaming.projectriff.io",
				Resource:  "kafkagatewaies",
				Namespace: defaultNamespace,
				Name:      kafkaGatewayName,
			}},
			ShouldError: true,
		},
	}

	table.Run(t, commands.NewKafkaGatewayDeleteCommand)
}
