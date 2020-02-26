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

	"github.com/projectriff/riff/cli/pkg/core/commands"
	rifftesting "github.com/projectriff/riff/cli/pkg/testing"
	"github.com/projectriff/system/pkg/apis"
	corev1alpha1 "github.com/projectriff/system/pkg/apis/core/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestDeployerStatusOptions(t *testing.T) {
	table := rifftesting.OptionsTable{
		{
			Name: "invalid resource",
			Options: &commands.DeployerStatusOptions{
				ResourceOptions: rifftesting.InvalidResourceOptions,
			},
			ExpectFieldErrors: rifftesting.InvalidResourceOptionsFieldError,
		},
		{
			Name: "valid resource",
			Options: &commands.DeployerStatusOptions{
				ResourceOptions: rifftesting.ValidResourceOptions,
			},
			ShouldValidate: true,
		},
	}

	table.Run(t)
}

func TestDeployerStatusCommand(t *testing.T) {
	defaultNamespace := "default"
	deployerName := "my-deployer"

	table := rifftesting.CommandTable{
		{
			Name:        "invalid args",
			Args:        []string{},
			ShouldError: true,
		},
		{
			Name: "show status",
			Args: []string{deployerName},
			GivenObjects: []runtime.Object{
				&corev1alpha1.Deployer{
					ObjectMeta: metav1.ObjectMeta{
						Name:      deployerName,
						Namespace: defaultNamespace,
					},
					Status: corev1alpha1.DeployerStatus{
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
# my-deployer: OopsieDoodle
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
			Args: []string{deployerName},
			ExpectOutput: `
Deployer "default/my-deployer" not found
`,
			ShouldError: true,
		},
		{
			Name: "get error",
			Args: []string{deployerName},
			GivenObjects: []runtime.Object{
				&corev1alpha1.Deployer{
					ObjectMeta: metav1.ObjectMeta{
						Name:      deployerName,
						Namespace: defaultNamespace,
					},
					Status: corev1alpha1.DeployerStatus{
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
				rifftesting.InduceFailure("get", "deployers"),
			},
			ShouldError: true,
		},
	}

	table.Run(t, commands.NewDeployerStatusCommand)
}
