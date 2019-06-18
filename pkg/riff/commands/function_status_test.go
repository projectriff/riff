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

func TestFunctionStatusOptions(t *testing.T) {
	table := rifftesting.OptionsTable{
		{
			Name: "invalid resource",
			Options: &commands.FunctionStatusOptions{
				ResourceOptions: rifftesting.InvalidResourceOptions,
			},
			ExpectFieldError: rifftesting.InvalidResourceOptionsFieldError,
		},
		{
			Name: "valid resource",
			Options: &commands.FunctionStatusOptions{
				ResourceOptions: rifftesting.ValidResourceOptions,
			},
			ShouldValidate: true,
		},
	}

	table.Run(t)
}

func TestFunctionStatusCommand(t *testing.T) {
	defaultNamespace := "default"
	functionName := "test-function"
	imageTag := "registry.example.com/repo:tag"
	gitRepo := "https://example.com/repo.git"
	gitMaster := "master"

	table := rifftesting.CommandTable{
		{
			Name:        "function does not exist",
			Args:        []string{functionName},
			ShouldError: true,
			ExpectOutput: `
functions.build.projectriff.io "test-function" not found
`,
		},
		{
			Name: "show status with ready status",
			Args: []string{functionName},
			GivenObjects: []runtime.Object{
				&buildv1alpha1.Function{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      functionName,
					},
					Spec: buildv1alpha1.FunctionSpec{
						Image: imageTag,
						Source: &buildv1alpha1.Source{
							Git: &buildv1alpha1.GitSource{
								URL:      gitRepo,
								Revision: gitMaster,
							},
						},
						Artifact: "uppercase.js",
						Handler:  "functions.Uppercase",
					},
					Status: buildv1alpha1.FunctionStatus{
						Status: duckv1alpha1.Status{
							Conditions: []duckv1alpha1.Condition{
								{
									Type:     buildv1alpha1.FunctionConditionReady,
									Status:   v1.ConditionTrue,
									Reason:   "",
									Severity: "",
									Message:  "",
								},
							},
						},
						BuildStatus: buildv1alpha1.BuildStatus{
							LatestImage: "projectriff/upper@sah256:abcdef1234",
						},
					},
				},
			},
			ExpectOutput: `
# test-function: Ready
---
lastTransitionTime: null
status: "True"
type: Ready
`,
		},
		{
			Name: "show status with non-ready status",
			Args: []string{functionName},
			GivenObjects: []runtime.Object{
				&buildv1alpha1.Function{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      functionName,
					},
					Spec: buildv1alpha1.FunctionSpec{
						Image: imageTag,
						Source: &buildv1alpha1.Source{
							Git: &buildv1alpha1.GitSource{
								URL:      gitRepo,
								Revision: gitMaster,
							},
						},
						Artifact: "uppercase.js",
						Handler:  "functions.Uppercase",
					},
					Status: buildv1alpha1.FunctionStatus{
						Status: duckv1alpha1.Status{
							Conditions: []duckv1alpha1.Condition{
								{
									Type:     buildv1alpha1.FunctionConditionReady,
									Status:   v1.ConditionFalse,
									Reason:   "",
									Severity: "",
									Message:  "build failed, check logs",
								},
							},
						},
						BuildStatus: buildv1alpha1.BuildStatus{},
					},
				},
			},
			ExpectOutput: `
# test-function: not-Ready
---
lastTransitionTime: null
message: build failed, check logs
status: "False"
type: Ready
`,
		},
	}

	table.Run(t, commands.NewFunctionStatusCommand)
}
