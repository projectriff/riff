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
	"time"

	"github.com/projectriff/riff/cli/pkg/streaming/commands"
	rifftesting "github.com/projectriff/riff/cli/pkg/testing"
	"github.com/projectriff/system/pkg/apis"
	streamv1alpha1 "github.com/projectriff/system/pkg/apis/streaming/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestPulsarGatewayStatusOptions(t *testing.T) {
	table := rifftesting.OptionsTable{
		{
			Name: "invalid resource",
			Options: &commands.PulsarGatewayStatusOptions{
				ResourceOptions: rifftesting.InvalidResourceOptions,
			},
			ExpectFieldErrors: rifftesting.InvalidResourceOptionsFieldError,
		},
		{
			Name: "valid resource",
			Options: &commands.PulsarGatewayStatusOptions{
				ResourceOptions: rifftesting.ValidResourceOptions,
			},
			ShouldValidate: true,
		},
	}

	table.Run(t)
}

func TestPulsarGatewayStatusCommand(t *testing.T) {
	defaultNamespace := "default"
	gatewayName := "my-gateway"

	table := rifftesting.CommandTable{
		{
			Name:        "invalid args",
			Args:        []string{},
			ShouldError: true,
		},
		{
			Name: "show status",
			Args: []string{gatewayName},
			GivenObjects: []runtime.Object{
				&streamv1alpha1.PulsarGateway{
					ObjectMeta: metav1.ObjectMeta{
						Name:      gatewayName,
						Namespace: defaultNamespace,
					},
					Status: streamv1alpha1.PulsarGatewayStatus{
						Status: apis.Status{
							Conditions: apis.Conditions{
								{
									Type:    apis.ConditionReady,
									Status:  corev1.ConditionFalse,
									Reason:  "OopsieDoodle",
									Message: "a hopefully informative message about what went wrong",
									LastTransitionTime: apis.VolatileTime{
										Inner: metav1.Time{
											Time: time.Date(2019, 6, 29, 01, 44, 05, 0, time.UTC),
										},
									},
								},
							},
						},
					},
				},
			},
			ExpectOutput: `
# my-gateway: OopsieDoodle
---
lastTransitionTime: "2019-06-29T01:44:05Z"
message: a hopefully informative message about what went wrong
reason: OopsieDoodle
status: "False"
type: Ready
`,
		},
		{
			Name: "not found",
			Args: []string{gatewayName},
			ExpectOutput: `
Pulsar gateway "default/my-gateway" not found
`,
			ShouldError: true,
		},
		{
			Name: "get error",
			Args: []string{gatewayName},
			GivenObjects: []runtime.Object{
				&streamv1alpha1.PulsarGateway{
					ObjectMeta: metav1.ObjectMeta{
						Name:      gatewayName,
						Namespace: defaultNamespace,
					},
					Status: streamv1alpha1.PulsarGatewayStatus{
						Status: apis.Status{
							Conditions: apis.Conditions{
								{
									Type:    apis.ConditionReady,
									Status:  corev1.ConditionFalse,
									Reason:  "OopsieDoodle",
									Message: "a hopefully informative message about what went wrong",
									LastTransitionTime: apis.VolatileTime{
										Inner: metav1.Time{
											Time: time.Date(2019, 6, 29, 01, 44, 05, 0, time.UTC),
										},
									},
								},
							},
						},
					},
				},
			},
			WithReactors: []rifftesting.ReactionFunc{
				rifftesting.InduceFailure("get", "pulsargatewaies"),
			},
			ShouldError: true,
		},
	}

	table.Run(t, commands.NewPulsarGatewayStatusCommand)
}
