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

func TestKafkaGatewayListOptions(t *testing.T) {
	table := rifftesting.OptionsTable{
		{
			Name: "invalid list",
			Options: &commands.KafkaGatewayListOptions{
				ListOptions: rifftesting.InvalidListOptions,
			},
			ExpectFieldErrors: rifftesting.InvalidListOptionsFieldError,
		},
		{
			Name: "valid list",
			Options: &commands.KafkaGatewayListOptions{
				ListOptions: rifftesting.ValidListOptions,
			},
			ShouldValidate: true,
		},
	}

	table.Run(t)
}

func TestKafkaGatewayListCommand(t *testing.T) {
	kafkaGatewayName := "test-kafka-gateway"
	kafkaGatewayOtherName := "test-other-kafka-gateway"
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
No kafka gateways found.
`,
		},
		{
			Name: "lists an item",
			Args: []string{},
			GivenObjects: []runtime.Object{
				&streamv1alpha1.KafkaGateway{
					ObjectMeta: metav1.ObjectMeta{
						Name:      kafkaGatewayName,
						Namespace: defaultNamespace,
					},
				},
			},
			ExpectOutput: `
NAME                 BOOTSTRAP SERVERS   STATUS      AGE
test-kafka-gateway   <empty>             <unknown>   <unknown>
`,
		},
		{
			Name: "filters by namespace",
			Args: []string{cli.NamespaceFlagName, otherNamespace},
			GivenObjects: []runtime.Object{
				&streamv1alpha1.KafkaGateway{
					ObjectMeta: metav1.ObjectMeta{
						Name:      kafkaGatewayName,
						Namespace: defaultNamespace,
					},
				},
			},
			ExpectOutput: `
No kafka gateways found.
`,
		},
		{
			Name: "all namespace",
			Args: []string{cli.AllNamespacesFlagName},
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
						Namespace: otherNamespace,
					},
				},
			},
			ExpectOutput: `
NAMESPACE         NAME                       BOOTSTRAP SERVERS   STATUS      AGE
default           test-kafka-gateway         <empty>             <unknown>   <unknown>
other-namespace   test-other-kafka-gateway   <empty>             <unknown>   <unknown>
`,
		},
		{
			Name: "table populates all columns",
			Args: []string{},
			GivenObjects: []runtime.Object{
				&streamv1alpha1.KafkaGateway{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "my-kafka",
						Namespace: defaultNamespace,
					},
					Spec: streamv1alpha1.KafkaGatewaySpec{
						BootstrapServers: "localhost:9092",
					},
					Status: streamv1alpha1.KafkaGatewayStatus{
						Status: apis.Status{
							Conditions: apis.Conditions{
								{Type: streamv1alpha1.KafkaGatewayConditionReady, Status: "True"},
							},
						},
					},
				},
			},
			ExpectOutput: `
NAME       BOOTSTRAP SERVERS   STATUS   AGE
my-kafka   localhost:9092      Ready    <unknown>
`,
		},
		{
			Name: "list error",
			Args: []string{},
			WithReactors: []rifftesting.ReactionFunc{
				rifftesting.InduceFailure("list", "kafkagatewaies"),
			},
			ShouldError: true,
		},
	}

	table.Run(t, commands.NewKafkaGatewayListCommand)
}
