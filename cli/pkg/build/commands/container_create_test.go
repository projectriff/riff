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
	"errors"
	"testing"

	"github.com/projectriff/cli/pkg/build/commands"
	"github.com/projectriff/cli/pkg/cli"
	"github.com/projectriff/cli/pkg/k8s"
	rifftesting "github.com/projectriff/cli/pkg/testing"
	buildv1alpha1 "github.com/projectriff/system/pkg/apis/build/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	cachetesting "k8s.io/client-go/tools/cache/testing"
)

func TestContainerCreateOptions(t *testing.T) {
	table := rifftesting.OptionsTable{
		{
			Name: "invalid resource",
			Options: &commands.ContainerCreateOptions{
				ResourceOptions: rifftesting.InvalidResourceOptions,
			},
			ExpectFieldErrors: rifftesting.InvalidResourceOptionsFieldError.Also(
				cli.ErrMissingField(cli.ImageFlagName),
			),
		},
		{
			Name: "image",
			Options: &commands.ContainerCreateOptions{
				ResourceOptions: rifftesting.ValidResourceOptions,
				Image:           "example.com/repo:tag",
			},
			ShouldValidate: true,
		},
		{
			Name: "tail",
			Options: &commands.ContainerCreateOptions{
				ResourceOptions: rifftesting.ValidResourceOptions,
				Image:           "example.com/repo:tag",
				Tail:            true,
				WaitTimeout:     "10m",
			},
			ShouldValidate: true,
		},
		{
			Name: "tail missing timeout",
			Options: &commands.ContainerCreateOptions{
				ResourceOptions: rifftesting.ValidResourceOptions,
				Image:           "example.com/repo:tag",
				Tail:            true,
			},
			ExpectFieldErrors: cli.ErrMissingField(cli.WaitTimeoutFlagName),
		},
		{
			Name: "tail invalid timeout",
			Options: &commands.ContainerCreateOptions{
				ResourceOptions: rifftesting.ValidResourceOptions,
				Image:           "example.com/repo:tag",
				Tail:            true,
				WaitTimeout:     "d",
			},
			ExpectFieldErrors: cli.ErrInvalidValue("d", cli.WaitTimeoutFlagName),
		},
		{
			Name: "dry run",
			Options: &commands.ContainerCreateOptions{
				ResourceOptions: rifftesting.ValidResourceOptions,
				Image:           "example.com/repo:tag",
				DryRun:          true,
			},
			ShouldValidate: true,
		},
		{
			Name: "dry run, tail",
			Options: &commands.ContainerCreateOptions{
				ResourceOptions: rifftesting.ValidResourceOptions,
				Image:           "example.com/repo:tag",
				Tail:            true,
				WaitTimeout:     "10m",
				DryRun:          true,
			},
			ExpectFieldErrors: cli.ErrMultipleOneOf(cli.DryRunFlagName, cli.TailFlagName),
		},
	}

	table.Run(t)
}

func TestContainerCreateCommand(t *testing.T) {
	defaultNamespace := "default"
	containerName := "my-container"
	imageDefault := "_"
	imageTag := "registry.example.com/repo:tag"

	table := rifftesting.CommandTable{
		{
			Name:        "invalid args",
			Args:        []string{},
			ShouldError: true,
		},
		{
			Name: "image",
			Args: []string{containerName, cli.ImageFlagName, imageTag},
			ExpectCreates: []runtime.Object{
				&buildv1alpha1.Container{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      containerName,
					},
					Spec: buildv1alpha1.ContainerSpec{
						Image: imageTag,
					},
				},
			},
			ExpectOutput: `
Created container "my-container"
`,
		},
		{
			Name: "default image",
			Args: []string{containerName},
			ExpectCreates: []runtime.Object{
				&buildv1alpha1.Container{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      containerName,
					},
					Spec: buildv1alpha1.ContainerSpec{
						Image: imageDefault,
					},
				},
			},
			ExpectOutput: `
Created container "my-container"
`,
		},
		{
			Name: "image dry run",
			Args: []string{containerName, cli.ImageFlagName, imageTag, cli.DryRunFlagName},
			ExpectOutput: `
---
apiVersion: build.projectriff.io/v1alpha1
kind: Container
metadata:
  creationTimestamp: null
  name: my-container
  namespace: default
spec:
  image: registry.example.com/repo:tag
status: {}

Created container "my-container"
`,
		},
		{
			Name: "error existing container",
			Args: []string{containerName, cli.ImageFlagName, imageTag},
			GivenObjects: []runtime.Object{
				&buildv1alpha1.Container{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      containerName,
					},
				},
			},
			ExpectCreates: []runtime.Object{
				&buildv1alpha1.Container{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      containerName,
					},
					Spec: buildv1alpha1.ContainerSpec{
						Image: imageTag,
					},
				},
			},
			ShouldError: true,
		},
		{
			Name: "error during create",
			Args: []string{containerName, cli.ImageFlagName, imageTag},
			WithReactors: []rifftesting.ReactionFunc{
				rifftesting.InduceFailure("create", "containers"),
			},
			ExpectCreates: []runtime.Object{
				&buildv1alpha1.Container{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      containerName,
					},
					Spec: buildv1alpha1.ContainerSpec{
						Image: imageTag,
					},
				},
			},
			ShouldError: true,
		},
		{
			Skip: true,
			Name: "tail logs",
			Args: []string{containerName, cli.ImageFlagName, imageTag, cli.TailFlagName},
			Prepare: func(t *testing.T, ctx context.Context, c *cli.Config) (context.Context, error) {
				lw := cachetesting.NewFakeControllerSource()
				ctx = k8s.WithListerWatcher(ctx, lw)

				return ctx, nil
			},
			CleanUp: func(t *testing.T, ctx context.Context, c *cli.Config) error {
				if lw, ok := k8s.GetListerWatcher(ctx, nil, "", nil).(*cachetesting.FakeControllerSource); ok {
					lw.Shutdown()
				}

				return nil
			},
			ExpectCreates: []runtime.Object{
				&buildv1alpha1.Container{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      containerName,
					},
					Spec: buildv1alpha1.ContainerSpec{
						Image: imageTag,
					},
				},
			},
			ExpectOutput: `
Created container "my-container"
Waiting for container "my-container" to become ready...
...log output...
Container "my-container" is ready
`,
		},
		{
			Name: "tail timeout",
			Args: []string{containerName, cli.ImageFlagName, imageTag, cli.TailFlagName, cli.WaitTimeoutFlagName, "5ms"},
			Prepare: func(t *testing.T, ctx context.Context, c *cli.Config) (context.Context, error) {
				lw := cachetesting.NewFakeControllerSource()
				ctx = k8s.WithListerWatcher(ctx, lw)

				return ctx, nil
			},
			CleanUp: func(t *testing.T, ctx context.Context, c *cli.Config) error {
				if lw, ok := k8s.GetListerWatcher(ctx, nil, "", nil).(*cachetesting.FakeControllerSource); ok {
					lw.Shutdown()
				}

				return nil
			},
			ExpectCreates: []runtime.Object{
				&buildv1alpha1.Container{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      containerName,
					},
					Spec: buildv1alpha1.ContainerSpec{
						Image: imageTag,
					},
				},
			},
			ExpectOutput: `
Created container "my-container"
Waiting for container "my-container" to become ready...
Timeout after "5ms" waiting for "my-container" to become ready
To view status run: riff container list --namespace default
To continue watching logs run: riff container tail my-container --namespace default
`,
			ShouldError: true,
			Verify: func(t *testing.T, output string, err error) {
				if actual := err; !errors.Is(err, cli.SilentError) {
					t.Errorf("expected error to be silent, actual %#v", actual)
				}
			},
		},
		{
			Skip: true,
			Name: "tail error",
			Args: []string{containerName, cli.ImageFlagName, imageTag, cli.TailFlagName},
			Prepare: func(t *testing.T, ctx context.Context, c *cli.Config) (context.Context, error) {
				lw := cachetesting.NewFakeControllerSource()
				ctx = k8s.WithListerWatcher(ctx, lw)

				return ctx, nil
			},
			CleanUp: func(t *testing.T, ctx context.Context, c *cli.Config) error {
				if lw, ok := k8s.GetListerWatcher(ctx, nil, "", nil).(*cachetesting.FakeControllerSource); ok {
					lw.Shutdown()
				}

				return nil
			},
			ExpectCreates: []runtime.Object{
				&buildv1alpha1.Container{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      containerName,
					},
					Spec: buildv1alpha1.ContainerSpec{
						Image: imageTag,
					},
				},
			},
			ExpectOutput: `
Created container "my-container"
Waiting for container "my-container" to become ready...
`,
			ShouldError: true,
		},
	}

	table.Run(t, commands.NewContainerCreateCommand)
}
