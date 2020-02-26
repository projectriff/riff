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

	"github.com/projectriff/riff/cli/pkg/cli"
	"github.com/projectriff/riff/cli/pkg/k8s"
	"github.com/projectriff/riff/cli/pkg/knative/commands"
	rifftesting "github.com/projectriff/riff/cli/pkg/testing"
	knativev1alpha1 "github.com/projectriff/system/pkg/apis/knative/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	cachetesting "k8s.io/client-go/tools/cache/testing"
)

func TestAdapterCreateOptions(t *testing.T) {
	table := rifftesting.OptionsTable{
		{
			Name: "invalid resource",
			Options: &commands.AdapterCreateOptions{
				ResourceOptions: rifftesting.InvalidResourceOptions,
			},
			ExpectFieldErrors: rifftesting.InvalidResourceOptionsFieldError.Also(
				cli.ErrMissingOneOf(cli.ApplicationRefFlagName, cli.ContainerRefFlagName, cli.FunctionRefFlagName),
				cli.ErrMissingOneOf(cli.ConfigurationRefFlagName, cli.ServiceRefFlagName),
			),
		},
		{
			Name: "from application",
			Options: &commands.AdapterCreateOptions{
				ResourceOptions: rifftesting.ValidResourceOptions,
				ApplicationRef:  "my-application",
				ServiceRef:      "my-service",
			},
			ShouldValidate: true,
		},
		{
			Name: "from container",
			Options: &commands.AdapterCreateOptions{
				ResourceOptions: rifftesting.ValidResourceOptions,
				ContainerRef:    "my-container",
				ServiceRef:      "my-service",
			},
			ShouldValidate: true,
		},
		{
			Name: "from function",
			Options: &commands.AdapterCreateOptions{
				ResourceOptions: rifftesting.ValidResourceOptions,
				FunctionRef:     "my-function",
				ServiceRef:      "my-service",
			},
			ShouldValidate: true,
		},
		{
			Name: "from application, container and function",
			Options: &commands.AdapterCreateOptions{
				ResourceOptions: rifftesting.ValidResourceOptions,
				ApplicationRef:  "my-application",
				ContainerRef:    "my-container",
				FunctionRef:     "my-function",
				ServiceRef:      "my-service",
			},
			ExpectFieldErrors: cli.ErrMultipleOneOf(cli.ApplicationRefFlagName, cli.ContainerRefFlagName, cli.FunctionRefFlagName),
		},
		{
			Name: "from service",
			Options: &commands.AdapterCreateOptions{
				ResourceOptions: rifftesting.ValidResourceOptions,
				ApplicationRef:  "my-application",
				ServiceRef:      "my-service",
			},
			ShouldValidate: true,
		},
		{
			Name: "from configuration",
			Options: &commands.AdapterCreateOptions{
				ResourceOptions:  rifftesting.ValidResourceOptions,
				ApplicationRef:   "my-application",
				ConfigurationRef: "my-configuration",
			},
			ShouldValidate: true,
		},
		{
			Name: "from service and configuration",
			Options: &commands.AdapterCreateOptions{
				ResourceOptions:  rifftesting.ValidResourceOptions,
				ApplicationRef:   "my-application",
				ConfigurationRef: "my-configuration",
				ServiceRef:       "my-service",
			},
			ExpectFieldErrors: cli.ErrMultipleOneOf(cli.ConfigurationRefFlagName, cli.ServiceRefFlagName),
		},
		{
			Name: "with tail",
			Options: &commands.AdapterCreateOptions{
				ResourceOptions: rifftesting.ValidResourceOptions,
				ApplicationRef:  "my-application",
				ServiceRef:      "my-service",
				Tail:            true,
				WaitTimeout:     "10m",
			},
			ShouldValidate: true,
		},
		{
			Name: "with tail, missing timeout",
			Options: &commands.AdapterCreateOptions{
				ResourceOptions: rifftesting.ValidResourceOptions,
				ApplicationRef:  "my-application",
				ServiceRef:      "my-service",
				Tail:            true,
			},
			ExpectFieldErrors: cli.ErrMissingField(cli.WaitTimeoutFlagName),
		},
		{
			Name: "with tail, invalid timeout",
			Options: &commands.AdapterCreateOptions{
				ResourceOptions: rifftesting.ValidResourceOptions,
				ApplicationRef:  "my-application",
				ServiceRef:      "my-service",
				Tail:            true,
				WaitTimeout:     "d",
			},
			ExpectFieldErrors: cli.ErrInvalidValue("d", cli.WaitTimeoutFlagName),
		},
		{
			Name: "dry run",
			Options: &commands.AdapterCreateOptions{
				ResourceOptions: rifftesting.ValidResourceOptions,
				ApplicationRef:  "my-application",
				ServiceRef:      "my-service",
				DryRun:          true,
			},
			ShouldValidate: true,
		},
		{
			Name: "dry run, tail",
			Options: &commands.AdapterCreateOptions{
				ResourceOptions: rifftesting.ValidResourceOptions,
				ApplicationRef:  "my-application",
				ServiceRef:      "my-service",
				Tail:            true,
				WaitTimeout:     "10m",
				DryRun:          true,
			},
			ExpectFieldErrors: cli.ErrMultipleOneOf(cli.DryRunFlagName, cli.TailFlagName),
		},
	}

	table.Run(t)
}

func TestAdapterCreateCommand(t *testing.T) {
	defaultNamespace := "default"
	adapterName := "my-adapter"
	applicationRef := "my-app"
	containerRef := "my-container"
	functionRef := "my-func"
	configurationRef := "my-config"
	serviceRef := "my-service"

	table := rifftesting.CommandTable{
		{
			Name:        "invalid args",
			Args:        []string{},
			ShouldError: true,
		},
		{
			Name: "create from application ref",
			Args: []string{adapterName, cli.ApplicationRefFlagName, applicationRef, cli.ServiceRefFlagName, serviceRef},
			ExpectCreates: []runtime.Object{
				&knativev1alpha1.Adapter{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      adapterName,
					},
					Spec: knativev1alpha1.AdapterSpec{
						Build: knativev1alpha1.Build{
							ApplicationRef: applicationRef,
						},
						Target: knativev1alpha1.AdapterTarget{
							ServiceRef: serviceRef,
						},
					},
				},
			},
			ExpectOutput: `
Created adapter "my-adapter"
`,
		},
		{
			Name: "create from container ref",
			Args: []string{adapterName, cli.ContainerRefFlagName, containerRef, cli.ServiceRefFlagName, serviceRef},
			ExpectCreates: []runtime.Object{
				&knativev1alpha1.Adapter{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      adapterName,
					},
					Spec: knativev1alpha1.AdapterSpec{
						Build: knativev1alpha1.Build{
							ContainerRef: containerRef,
						},
						Target: knativev1alpha1.AdapterTarget{
							ServiceRef: serviceRef,
						},
					},
				},
			},
			ExpectOutput: `
Created adapter "my-adapter"
`,
		},
		{
			Name: "create from function ref",
			Args: []string{adapterName, cli.FunctionRefFlagName, functionRef, cli.ServiceRefFlagName, serviceRef},
			ExpectCreates: []runtime.Object{
				&knativev1alpha1.Adapter{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      adapterName,
					},
					Spec: knativev1alpha1.AdapterSpec{
						Build: knativev1alpha1.Build{
							FunctionRef: functionRef,
						},
						Target: knativev1alpha1.AdapterTarget{
							ServiceRef: serviceRef,
						},
					},
				},
			},
			ExpectOutput: `
Created adapter "my-adapter"
`,
		},
		{
			Name: "create from configuration ref",
			Args: []string{adapterName, cli.FunctionRefFlagName, functionRef, cli.ConfigurationRefFlagName, configurationRef},
			ExpectCreates: []runtime.Object{
				&knativev1alpha1.Adapter{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      adapterName,
					},
					Spec: knativev1alpha1.AdapterSpec{
						Build: knativev1alpha1.Build{
							FunctionRef: functionRef,
						},
						Target: knativev1alpha1.AdapterTarget{
							ConfigurationRef: configurationRef,
						},
					},
				},
			},
			ExpectOutput: `
Created adapter "my-adapter"
`,
		},
		{
			Name: "create from service ref",
			Args: []string{adapterName, cli.FunctionRefFlagName, functionRef, cli.ServiceRefFlagName, serviceRef},
			ExpectCreates: []runtime.Object{
				&knativev1alpha1.Adapter{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      adapterName,
					},
					Spec: knativev1alpha1.AdapterSpec{
						Build: knativev1alpha1.Build{
							FunctionRef: functionRef,
						},
						Target: knativev1alpha1.AdapterTarget{
							ServiceRef: serviceRef,
						},
					},
				},
			},
			ExpectOutput: `
Created adapter "my-adapter"
`,
		},
		{
			Name: "dry run",
			Args: []string{adapterName, cli.FunctionRefFlagName, functionRef, cli.ServiceRefFlagName, serviceRef, cli.DryRunFlagName},
			ExpectOutput: `
---
apiVersion: knative.projectriff.io/v1alpha1
kind: Adapter
metadata:
  creationTimestamp: null
  name: my-adapter
  namespace: default
spec:
  build:
    functionRef: my-func
  target:
    serviceRef: my-service
status: {}

Created adapter "my-adapter"
`,
		},
		{
			Name: "error existing adapter",
			Args: []string{adapterName, cli.FunctionRefFlagName, functionRef, cli.ServiceRefFlagName, serviceRef},
			GivenObjects: []runtime.Object{
				&knativev1alpha1.Adapter{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      adapterName,
					},
				},
			},
			ExpectCreates: []runtime.Object{
				&knativev1alpha1.Adapter{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      adapterName,
					},
					Spec: knativev1alpha1.AdapterSpec{
						Build: knativev1alpha1.Build{
							FunctionRef: functionRef,
						},
						Target: knativev1alpha1.AdapterTarget{
							ServiceRef: serviceRef,
						},
					},
				},
			},
			ShouldError: true,
		},
		{
			Name: "error during create",
			Args: []string{adapterName, cli.FunctionRefFlagName, functionRef, cli.ServiceRefFlagName, serviceRef},
			WithReactors: []rifftesting.ReactionFunc{
				rifftesting.InduceFailure("create", "adapters"),
			},
			ExpectCreates: []runtime.Object{
				&knativev1alpha1.Adapter{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      adapterName,
					},
					Spec: knativev1alpha1.AdapterSpec{
						Build: knativev1alpha1.Build{
							FunctionRef: functionRef,
						},
						Target: knativev1alpha1.AdapterTarget{
							ServiceRef: serviceRef,
						},
					},
				},
			},
			ShouldError: true,
		},
		{
			Skip: true,
			Name: "tail logs",
			Args: []string{adapterName, cli.FunctionRefFlagName, functionRef, cli.ServiceRefFlagName, serviceRef, cli.TailFlagName},
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
				&knativev1alpha1.Adapter{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      adapterName,
					},
					Spec: knativev1alpha1.AdapterSpec{
						Build: knativev1alpha1.Build{
							FunctionRef: functionRef,
						},
						Target: knativev1alpha1.AdapterTarget{
							ServiceRef: serviceRef,
						},
					},
				},
			},
			ExpectOutput: `
Created adapter "my-adapter"
Waiting for adapter "my-adapter" to become ready...
...log output...
Adapter "my-adapter" is ready
`,
		},
		{
			Name: "tail timeout",
			Args: []string{adapterName, cli.FunctionRefFlagName, functionRef, cli.ServiceRefFlagName, serviceRef, cli.TailFlagName, cli.WaitTimeoutFlagName, "5ms"},
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
				&knativev1alpha1.Adapter{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      adapterName,
					},
					Spec: knativev1alpha1.AdapterSpec{
						Build: knativev1alpha1.Build{
							FunctionRef: functionRef,
						},
						Target: knativev1alpha1.AdapterTarget{
							ServiceRef: serviceRef,
						},
					},
				},
			},
			ExpectOutput: `
Created adapter "my-adapter"
Waiting for adapter "my-adapter" to become ready...
Timeout after "5ms" waiting for "my-adapter" to become ready
To view status run: riff knative adapter list --namespace default
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
			Args: []string{adapterName, cli.FunctionRefFlagName, functionRef, cli.ServiceRefFlagName, serviceRef, cli.TailFlagName},
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
				&knativev1alpha1.Adapter{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      adapterName,
					},
					Spec: knativev1alpha1.AdapterSpec{
						Build: knativev1alpha1.Build{
							FunctionRef: functionRef,
						},
						Target: knativev1alpha1.AdapterTarget{
							ServiceRef: serviceRef,
						},
					},
				},
			},
			ExpectOutput: `
Created adapter "my-adapter"
Waiting for adapter "my-adapter" to become ready...
`,
			ShouldError: true,
		},
	}

	table.Run(t, commands.NewAdapterCreateCommand)
}
