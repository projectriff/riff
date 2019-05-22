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
	"testing"

	"github.com/projectriff/riff/pkg/cli"
	"github.com/projectriff/riff/pkg/riff/commands"
	rifftesting "github.com/projectriff/riff/pkg/testing"
	requestv1alpha1 "github.com/projectriff/system/pkg/apis/request/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestRouteCreateOptions(t *testing.T) {
	table := rifftesting.OptionsTable{
		{
			Name: "invalid resource",
			Options: &commands.RouteCreateOptions{
				ResourceOptions: rifftesting.InvalidResourceOptions,
			},
			ExpectFieldError: rifftesting.InvalidResourceOptionsFieldError.Also(
				cli.ErrMissingOneOf(cli.ApplicationRefFlagName, cli.FunctionRefFlagName, cli.ImageFlagName),
			),
		},
		{
			Name: "from application",
			Options: &commands.RouteCreateOptions{
				ResourceOptions: rifftesting.ValidResourceOptions,
				ApplicationRef:  "my-application",
			},
			ShouldValidate: true,
		},
		{
			Name: "from function",
			Options: &commands.RouteCreateOptions{
				ResourceOptions: rifftesting.ValidResourceOptions,
				FunctionRef:     "my-function",
			},
			ShouldValidate: true,
		},
		{
			Name: "from image",
			Options: &commands.RouteCreateOptions{
				ResourceOptions: rifftesting.ValidResourceOptions,
				Image:           "example.com/repo:tag",
			},
			ShouldValidate: true,
		},
		{
			Name: "from application, funcation and image",
			Options: &commands.RouteCreateOptions{
				ResourceOptions: rifftesting.ValidResourceOptions,
				ApplicationRef:  "my-application",
				FunctionRef:     "my-function",
				Image:           "example.com/repo:tag",
			},
			ExpectFieldError: cli.ErrMultipleOneOf(cli.ApplicationRefFlagName, cli.FunctionRefFlagName, cli.ImageFlagName),
		},
		{
			Name: "with env",
			Options: &commands.RouteCreateOptions{
				ResourceOptions: rifftesting.ValidResourceOptions,
				Image:           "example.com/repo:tag",
				Env:             []string{"VAR1=foo", "VAR2=bar"},
			},
			ShouldValidate: true,
		},
		{
			Name: "with invalid env",
			Options: &commands.RouteCreateOptions{
				ResourceOptions: rifftesting.ValidResourceOptions,
				Image:           "example.com/repo:tag",
				Env:             []string{"=foo"},
			},
			ExpectFieldError: cli.ErrInvalidArrayValue("=foo", cli.EnvFlagName, 0),
		},
	}

	table.Run(t)
}

func TestRouteCreateCommand(t *testing.T) {
	defaultNamespace := "default"
	routeName := "my-route"
	image := "registry.example.com/repo@sha256:deadbeefdeadbeefdeadbeefdeadbeef"
	applicationRef := "my-app"
	functionRef := "my-func"
	envName := "MY_VAR"
	envValue := "my-value"
	envVar := fmt.Sprintf("%s=%s", envName, envValue)
	envNameOther := "MY_VAR_OTHER"
	envValueOther := "my-value-other"
	envVarOther := fmt.Sprintf("%s=%s", envNameOther, envValueOther)

	table := rifftesting.CommandTable{
		{
			Name:        "invalid args",
			Args:        []string{},
			ShouldError: true,
		},
		{
			Name: "create from image",
			Args: []string{routeName, cli.ImageFlagName, image},
			ExpectCreates: []runtime.Object{
				&requestv1alpha1.Route{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      routeName,
					},
					Spec: requestv1alpha1.RouteSpec{
						Template: &corev1.PodSpec{
							Containers: []corev1.Container{
								{Image: image},
							},
						},
					},
				},
			},
			ExpectOutput: `
Created route "my-route"
`,
		},
		{
			Name: "create from application ref",
			Args: []string{routeName, cli.ApplicationRefFlagName, applicationRef},
			ExpectCreates: []runtime.Object{
				&requestv1alpha1.Route{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      routeName,
					},
					Spec: requestv1alpha1.RouteSpec{
						Build: &requestv1alpha1.Build{
							ApplicationRef: applicationRef,
						},
					},
				},
			},
			ExpectOutput: `
Created route "my-route"
`,
		},
		{
			Name: "create from function ref",
			Args: []string{routeName, cli.FunctionRefFlagName, functionRef},
			ExpectCreates: []runtime.Object{
				&requestv1alpha1.Route{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      routeName,
					},
					Spec: requestv1alpha1.RouteSpec{
						Build: &requestv1alpha1.Build{
							FunctionRef: functionRef,
						},
					},
				},
			},
			ExpectOutput: `
Created route "my-route"
`,
		},
		{
			Name: "create from image with env",
			Args: []string{routeName, cli.ImageFlagName, image, cli.EnvFlagName, envVar, cli.EnvFlagName, envVarOther},
			ExpectCreates: []runtime.Object{
				&requestv1alpha1.Route{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      routeName,
					},
					Spec: requestv1alpha1.RouteSpec{
						Template: &corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Image: image,
									Env: []corev1.EnvVar{
										{Name: envName, Value: envValue},
										{Name: envNameOther, Value: envValueOther},
									},
								},
							},
						},
					},
				},
			},
			ExpectOutput: `
Created route "my-route"
`,
		},
		{
			Name: "error existing route",
			Args: []string{routeName, cli.ImageFlagName, image},
			GivenObjects: []runtime.Object{
				&requestv1alpha1.Route{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      routeName,
					},
				},
			},
			ExpectCreates: []runtime.Object{
				&requestv1alpha1.Route{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      routeName,
					},
					Spec: requestv1alpha1.RouteSpec{
						Template: &corev1.PodSpec{
							Containers: []corev1.Container{
								{Image: image},
							},
						},
					},
				},
			},
			ShouldError: true,
		},
		{
			Name: "error during create",
			Args: []string{routeName, cli.ImageFlagName, image},
			WithReactors: []rifftesting.ReactionFunc{
				rifftesting.InduceFailure("create", "routes"),
			},
			ExpectCreates: []runtime.Object{
				&requestv1alpha1.Route{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      routeName,
					},
					Spec: requestv1alpha1.RouteSpec{
						Template: &corev1.PodSpec{
							Containers: []corev1.Container{
								{Image: image},
							},
						},
					},
				},
			},
			ShouldError: true,
		},
	}

	table.Run(t, commands.NewRouteCreateCommand)
}
