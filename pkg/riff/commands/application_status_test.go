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

	duckv1alpha1 "github.com/knative/pkg/apis/duck/v1alpha1"
	buildv1alpha1 "github.com/projectriff/system/pkg/apis/build/v1alpha1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/projectriff/riff/pkg/riff/commands"
	rifftesting "github.com/projectriff/riff/pkg/testing"
)

func TestApplicationStatusOptions(t *testing.T) {
	table := rifftesting.OptionsTable{
		{
			Name: "invalid resource",
			Options: &commands.ApplicationStatusOptions{
				ResourceOptions: rifftesting.InvalidResourceOptions,
			},
			ExpectFieldError: rifftesting.InvalidResourceOptionsFieldError,
		},
		{
			Name: "valid resource",
			Options: &commands.ApplicationStatusOptions{
				ResourceOptions: rifftesting.ValidResourceOptions,
			},
			ShouldValidate: true,
		},
	}

	table.Run(t)
}

func TestApplicationStatusCommand(t *testing.T) {
	defaultNamespace := "default"
	appName := "test-app"
	imageTag := "registry.example.com/repo:tag"
	gitRepo := "https://example.com/repo.git"
	gitMaster := "master"

	table := rifftesting.CommandTable{
		{
			Name:        "app does not exist",
			Args:        []string{appName},
			ShouldError: true,
			ExpectOutput: `
Application "default/test-app" not found
`,
		},
		{
			Name: "handle error",
			Args: []string{appName},
			GivenObjects: []runtime.Object{
				&buildv1alpha1.Application{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      appName,
					},
				},
			},
			WithReactors: []rifftesting.ReactionFunc{
				rifftesting.InduceFailure("get", "applications"),
			},
			ShouldError: true,
		},
		{
			Name: "show status with ready status",
			Args: []string{appName},
			GivenObjects: []runtime.Object{
				&buildv1alpha1.Application{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      appName,
					},
					Spec: buildv1alpha1.ApplicationSpec{
						Image: imageTag,
						Source: &buildv1alpha1.Source{
							Git: &buildv1alpha1.GitSource{
								URL:      gitRepo,
								Revision: gitMaster,
							},
						},
					},
					Status: buildv1alpha1.ApplicationStatus{
						Status: duckv1alpha1.Status{
							Conditions: []duckv1alpha1.Condition{
								{
									Type:     buildv1alpha1.ApplicationConditionReady,
									Status:   v1.ConditionTrue,
								},
							},
						},
						BuildStatus: buildv1alpha1.BuildStatus{
							LatestImage: "projectriff/myapp@sah256:abcdef1234",
						},
					},
				},
			},
			ExpectOutput: `
# test-app: Ready
---
lastTransitionTime: null
status: "True"
type: Ready
`,
		},
		{
			Name: "show status with non-ready status and reason",
			Args: []string{appName},
			GivenObjects: []runtime.Object{
				&buildv1alpha1.Application{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      appName,
					},
					Spec: buildv1alpha1.ApplicationSpec{
						Image: imageTag,
						Source: &buildv1alpha1.Source{
							Git: &buildv1alpha1.GitSource{
								URL:      gitRepo,
								Revision: gitMaster,
							},
						},
					},
					Status: buildv1alpha1.ApplicationStatus{
						Status: duckv1alpha1.Status{
							Conditions: []duckv1alpha1.Condition{
								{
									Type:     buildv1alpha1.ApplicationConditionReady,
									Status:   v1.ConditionFalse,
									Reason:   "Failure",
									Severity: "Severe",
									Message:  "build failed, check logs",
								},
							},
						},
						BuildStatus: buildv1alpha1.BuildStatus{},
					},
				},
			},
			ExpectOutput: `
# test-app: Failure
---
lastTransitionTime: null
message: build failed, check logs
reason: Failure
severity: Severe
status: "False"
type: Ready
`,
		},
		{
			Name: "show status with non-ready status and only message",
			Args: []string{appName},
			GivenObjects: []runtime.Object{
				&buildv1alpha1.Application{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      appName,
					},
					Spec: buildv1alpha1.ApplicationSpec{
						Image: imageTag,
						Source: &buildv1alpha1.Source{
							Git: &buildv1alpha1.GitSource{
								URL:      gitRepo,
								Revision: gitMaster,
							},
						},
					},
					Status: buildv1alpha1.ApplicationStatus{
						Status: duckv1alpha1.Status{
							Conditions: []duckv1alpha1.Condition{
								{
									Type:     buildv1alpha1.ApplicationConditionReady,
									Status:   v1.ConditionFalse,
									Message:  "build failed, check logs",
								},
							},
						},
						BuildStatus: buildv1alpha1.BuildStatus{},
					},
				},
			},
			ExpectOutput: `
# test-app: not-Ready
---
lastTransitionTime: null
message: build failed, check logs
status: "False"
type: Ready
`,
		},
	}

	table.Run(t, commands.NewApplicationStatusCommand)
}
