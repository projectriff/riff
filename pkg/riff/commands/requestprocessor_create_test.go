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
	"fmt"

	"github.com/projectriff/riff/pkg/riff/commands"
	"github.com/projectriff/riff/pkg/testing"
	requestv1alpha1 "github.com/projectriff/system/pkg/apis/request/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestRequestProcessorCreateCommand(t *testing.T) {
	t.Parallel()

	defaultNamespace := "default"
	requestProcessorName := "my-function"
	itemName := "blue"
	image := "registry.example.com/repo@sha256:deadbeefdeadbeefdeadbeefdeadbeef"
	applicationRef := "my-app"
	functionRef := "my-func"
	envName := "MY_VAR"
	envValue := "my-value"
	envVar := fmt.Sprintf("%s=%s", envName, envValue)

	table := testing.CommandTable{
		{
			Name:        "empty",
			Args:        []string{},
			ShouldError: true,
		},
		{
			Name: "create from image",
			Args: []string{requestProcessorName, "--item", itemName, "--image", image},
			ExpectCreates: []runtime.Object{
				&requestv1alpha1.RequestProcessor{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      requestProcessorName,
					},
					Spec: requestv1alpha1.RequestProcessorSpec{
						{
							Name: itemName,
							Template: &corev1.PodSpec{
								Containers: []corev1.Container{
									{Image: image},
								},
							},
						},
					},
				},
			},
		},
		{
			Name: "create from application ref",
			Args: []string{requestProcessorName, "--item", itemName, "--application-ref", applicationRef},
			ExpectCreates: []runtime.Object{
				&requestv1alpha1.RequestProcessor{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      requestProcessorName,
					},
					Spec: requestv1alpha1.RequestProcessorSpec{
						{
							Name: itemName,
							Build: &requestv1alpha1.Build{
								ApplicationRef: applicationRef,
							},
						},
					},
				},
			},
		},
		{
			Name: "create from function ref",
			Args: []string{requestProcessorName, "--item", itemName, "--function-ref", functionRef},
			ExpectCreates: []runtime.Object{
				&requestv1alpha1.RequestProcessor{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      requestProcessorName,
					},
					Spec: requestv1alpha1.RequestProcessorSpec{
						{
							Name: itemName,
							Build: &requestv1alpha1.Build{
								FunctionRef: functionRef,
							},
						},
					},
				},
			},
		},
		{
			Name: "create from image with env",
			Args: []string{requestProcessorName, "--item", itemName, "--image", image, "--env", envVar},
			ExpectCreates: []runtime.Object{
				&requestv1alpha1.RequestProcessor{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      requestProcessorName,
					},
					Spec: requestv1alpha1.RequestProcessorSpec{
						{
							Name: itemName,
							Template: &corev1.PodSpec{
								Containers: []corev1.Container{
									{
										Image: image,
										Env: []corev1.EnvVar{
											{Name: envName, Value: envValue},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			// TODO impelement
			Skip: true,
			Name: "create from image with env from",
			Args: []string{requestProcessorName, "--item", itemName, "--image", image, "--env-from", "<todo>"},
			ExpectCreates: []runtime.Object{
				&requestv1alpha1.RequestProcessor{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      requestProcessorName,
					},
					Spec: requestv1alpha1.RequestProcessorSpec{
						{
							Name: itemName,
							Template: &corev1.PodSpec{
								Containers: []corev1.Container{
									{
										Image: image,
										Env: []corev1.EnvVar{
											{
												Name:      envName,
												ValueFrom: &corev1.EnvVarSource{
													// TODO implement
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			Name: "error existing request processor",
			Args: []string{requestProcessorName, "--item", itemName, "--image", image},
			GivenObjects: []runtime.Object{
				&requestv1alpha1.RequestProcessor{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      requestProcessorName,
					},
				},
			},
			ExpectCreates: []runtime.Object{
				&requestv1alpha1.RequestProcessor{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      requestProcessorName,
					},
					Spec: requestv1alpha1.RequestProcessorSpec{
						{
							Name: itemName,
							Template: &corev1.PodSpec{
								Containers: []corev1.Container{
									{Image: image},
								},
							},
						},
					},
				},
			},
			ShouldError: true,
		},
		{
			Name: "error durring create",
			Args: []string{requestProcessorName, "--item", itemName, "--image", image},
			WithReactors: []testing.ReactionFunc{
				testing.InduceFailure("create", "requestprocessors"),
			},
			ExpectCreates: []runtime.Object{
				&requestv1alpha1.RequestProcessor{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      requestProcessorName,
					},
					Spec: requestv1alpha1.RequestProcessorSpec{
						{
							Name: itemName,
							Template: &corev1.PodSpec{
								Containers: []corev1.Container{
									{Image: image},
								},
							},
						},
					},
				},
			},
			ShouldError: true,
		},
	}

	table.Run(t, commands.NewRequestProcessorCreateCommand)
}
