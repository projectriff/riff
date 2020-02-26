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
	"fmt"
	"testing"

	"github.com/projectriff/cli/pkg/cli"
	"github.com/projectriff/cli/pkg/k8s"
	"github.com/projectriff/cli/pkg/streaming/commands"
	rifftesting "github.com/projectriff/cli/pkg/testing"
	kailtesting "github.com/projectriff/cli/pkg/testing/kail"
	streamingv1alpha1 "github.com/projectriff/system/pkg/apis/streaming/v1alpha1"
	"github.com/stretchr/testify/mock"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	cachetesting "k8s.io/client-go/tools/cache/testing"
)

func TestProcessorCreateOptions(t *testing.T) {
	table := rifftesting.OptionsTable{
		{
			Name: "invalid resource",
			Options: &commands.ProcessorCreateOptions{
				ResourceOptions: rifftesting.InvalidResourceOptions,
			},
			ExpectFieldErrors: rifftesting.InvalidResourceOptionsFieldError.Also(
				cli.ErrMissingOneOf(cli.ContainerRefFlagName, cli.FunctionRefFlagName, cli.ImageFlagName),
				cli.ErrMissingField(cli.InputFlagName),
			),
		},
		{
			Name: "with inputs but no outputs",
			Options: &commands.ProcessorCreateOptions{
				ResourceOptions: rifftesting.ValidResourceOptions,
				FunctionRef:     "my-function",
				Inputs:          []string{"input1", "input2"},
			},
			ShouldValidate: true,
		}, {
			Name: "with image",
			Options: &commands.ProcessorCreateOptions{
				ResourceOptions: rifftesting.ValidResourceOptions,
				Image:           "my-image",
				Inputs:          []string{"input"},
			},
			ShouldValidate: true,
		}, {
			Name: "with container ref",
			Options: &commands.ProcessorCreateOptions{
				ResourceOptions: rifftesting.ValidResourceOptions,
				ContainerRef:    "my-container",
				Inputs:          []string{"input"},
			},
			ShouldValidate: true,
		}, {
			Name: "with function ref",
			Options: &commands.ProcessorCreateOptions{
				ResourceOptions: rifftesting.ValidResourceOptions,
				FunctionRef:     "my-function",
				Inputs:          []string{"input"},
			},
			ShouldValidate: true,
		}, {
			Name: "invalid with image, container ref and function ref",
			Options: &commands.ProcessorCreateOptions{
				ResourceOptions: rifftesting.ValidResourceOptions,
				Image:           "my-image",
				ContainerRef:    "my-container",
				FunctionRef:     "my-function",
				Inputs:          []string{"input1", "input2"},
			},
			ExpectFieldErrors: cli.ErrMultipleOneOf(cli.ContainerRefFlagName, cli.FunctionRefFlagName, cli.ImageFlagName),
		},
		{
			Name: "with inputs and outputs",
			Options: &commands.ProcessorCreateOptions{
				ResourceOptions: rifftesting.ValidResourceOptions,
				FunctionRef:     "my-function",
				Inputs:          []string{"input1", "input2"},
				Outputs:         []string{"output1", "output2"},
			},
			ShouldValidate: true,
		},
		{
			Name: "with explicit input bindings and output bindings",
			Options: &commands.ProcessorCreateOptions{
				ResourceOptions: rifftesting.ValidResourceOptions,
				FunctionRef:     "my-function",
				Inputs:          []string{"inParam1:input1", "inParam2:input2"},
				Outputs:         []string{"outParam1:output1", "outParam1:output2"},
			},
			ShouldValidate: true,
		},
		{
			Name: "with env",
			Options: &commands.ProcessorCreateOptions{
				ResourceOptions: rifftesting.ValidResourceOptions,
				Image:           "example.com/repo:tag",
				Inputs:          []string{"input1", "input2"},
				Env:             []string{"VAR1=foo", "VAR2=bar"},
			},
			ShouldValidate: true,
		},
		{
			Name: "with invalid env",
			Options: &commands.ProcessorCreateOptions{
				ResourceOptions: rifftesting.ValidResourceOptions,
				Image:           "example.com/repo:tag",
				Inputs:          []string{"input1", "input2"},
				Env:             []string{"=foo"},
			},
			ExpectFieldErrors: cli.ErrInvalidArrayValue("=foo", cli.EnvFlagName, 0),
		},
		{
			Name: "with envfrom secret",
			Options: &commands.ProcessorCreateOptions{
				ResourceOptions: rifftesting.ValidResourceOptions,
				Image:           "example.com/repo:tag",
				Inputs:          []string{"input1", "input2"},
				EnvFrom:         []string{"VAR1=secretKeyRef:name:key"},
			},
			ShouldValidate: true,
		},
		{
			Name: "with envfrom configmap",
			Options: &commands.ProcessorCreateOptions{
				ResourceOptions: rifftesting.ValidResourceOptions,
				Image:           "example.com/repo:tag",
				Inputs:          []string{"input1", "input2"},
				EnvFrom:         []string{"VAR1=configMapKeyRef:name:key"},
			},
			ShouldValidate: true,
		},
		{
			Name: "with invalid envfrom",
			Options: &commands.ProcessorCreateOptions{
				ResourceOptions: rifftesting.ValidResourceOptions,
				Image:           "example.com/repo:tag",
				Inputs:          []string{"input1", "input2"},
				EnvFrom:         []string{"VAR1=someOtherKeyRef:name:key"},
			},
			ExpectFieldErrors: cli.ErrInvalidArrayValue("VAR1=someOtherKeyRef:name:key", cli.EnvFromFlagName, 0),
		},
		{
			Name: "with tail",
			Options: &commands.ProcessorCreateOptions{
				ResourceOptions: rifftesting.ValidResourceOptions,
				FunctionRef:     "my-function",
				Inputs:          []string{"input"},
				Tail:            true,
				WaitTimeout:     "10m",
			},
			ShouldValidate: true,
		},
		{
			Name: "with tail, missing timeout",
			Options: &commands.ProcessorCreateOptions{
				ResourceOptions: rifftesting.ValidResourceOptions,
				FunctionRef:     "my-function",
				Inputs:          []string{"input"},
				Tail:            true,
			},
			ExpectFieldErrors: cli.ErrMissingField(cli.WaitTimeoutFlagName),
		},
		{
			Name: "with tail, invalid timeout",
			Options: &commands.ProcessorCreateOptions{
				ResourceOptions: rifftesting.ValidResourceOptions,
				FunctionRef:     "my-function",
				Inputs:          []string{"input"},
				Tail:            true,
				WaitTimeout:     "d",
			},
			ExpectFieldErrors: cli.ErrInvalidValue("d", cli.WaitTimeoutFlagName),
		},
		{
			Name: "dry run",
			Options: &commands.ProcessorCreateOptions{
				ResourceOptions: rifftesting.ValidResourceOptions,
				FunctionRef:     "my-function",
				Inputs:          []string{"input"},
				DryRun:          true,
			},
			ShouldValidate: true,
		},
		{
			Name: "dry run, tail",
			Options: &commands.ProcessorCreateOptions{
				ResourceOptions: rifftesting.ValidResourceOptions,
				FunctionRef:     "my-function",
				Inputs:          []string{"input"},
				Tail:            true,
				WaitTimeout:     "10m",
				DryRun:          true,
			},
			ExpectFieldErrors: cli.ErrMultipleOneOf(cli.DryRunFlagName, cli.TailFlagName),
		},
	}

	table.Run(t)
}

func TestProcessorCreateCommand(t *testing.T) {
	defaultNamespace := "default"
	processorName := "my-processor"
	containerRef := "my-container"
	functionRef := "my-func"
	image := "my-image"
	inputName := "input"
	inParameterName := "in"
	inputNameBinding := fmt.Sprintf("%s:%s", inParameterName, inputName)
	outputName := "output"
	outParameterName := "out"
	outputNameBinding := fmt.Sprintf("%s:%s", outParameterName, outputName)
	inputNameOther := "otherinput"
	outputNameOther := "otheroutput"
	envName := "MY_VAR"
	envValue := "my-value"
	envVar := fmt.Sprintf("%s=%s", envName, envValue)
	envNameOther := "MY_VAR_OTHER"
	envValueOther := "my-value-other"
	envVarOther := fmt.Sprintf("%s=%s", envNameOther, envValueOther)
	envVarFromConfigMap := "MY_VAR_FROM_CONFIGMAP=configMapKeyRef:my-configmap:my-key"
	envVarFromSecret := "MY_VAR_FROM_SECRET=secretKeyRef:my-secret:my-key"

	table := rifftesting.CommandTable{
		{
			Name:        "invalid args",
			Args:        []string{},
			ShouldError: true,
		},
		{
			Name: "create with container ref",
			Args: []string{processorName, cli.ContainerRefFlagName, containerRef, cli.InputFlagName, inputName},
			ExpectCreates: []runtime.Object{
				&streamingv1alpha1.Processor{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      processorName,
					},
					Spec: streamingv1alpha1.ProcessorSpec{
						Build:  &streamingv1alpha1.Build{ContainerRef: containerRef},
						Inputs: []streamingv1alpha1.InputStreamBinding{{Stream: inputName}},
						Template: &corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{},
								},
							},
						},
					},
				},
			},
			ExpectOutput: `
Created processor "my-processor"
`,
		},
		{
			Name: "create with function ref",
			Args: []string{processorName, cli.FunctionRefFlagName, functionRef, cli.InputFlagName, inputName},
			ExpectCreates: []runtime.Object{
				&streamingv1alpha1.Processor{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      processorName,
					},
					Spec: streamingv1alpha1.ProcessorSpec{
						Build:  &streamingv1alpha1.Build{FunctionRef: functionRef},
						Inputs: []streamingv1alpha1.InputStreamBinding{{Stream: inputName}},
						Template: &corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{},
								},
							},
						},
					},
				},
			},
			ExpectOutput: `
Created processor "my-processor"
`,
		},
		{
			Name: "create with image",
			Args: []string{processorName, cli.ImageFlagName, image, cli.InputFlagName, inputName},
			ExpectCreates: []runtime.Object{
				&streamingv1alpha1.Processor{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      processorName,
					},
					Spec: streamingv1alpha1.ProcessorSpec{
						Inputs: []streamingv1alpha1.InputStreamBinding{{Stream: inputName}},
						Template: &corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{Image: image},
								},
							},
						},
					},
				},
			},
			ExpectOutput: `
Created processor "my-processor"
`,
		},
		{
			Name: "dry run",
			Args: []string{processorName, cli.FunctionRefFlagName, functionRef, cli.InputFlagName, inputName, cli.DryRunFlagName},
			ExpectOutput: `
---
apiVersion: streaming.projectriff.io/v1alpha1
kind: Processor
metadata:
  creationTimestamp: null
  name: my-processor
  namespace: default
spec:
  build:
    functionRef: my-func
  inputs:
  - startOffset: ""
    stream: input
  outputs: []
  template:
    metadata:
      creationTimestamp: null
    spec:
      containers:
      - name: ""
        resources: {}
status: {}

Created processor "my-processor"
`,
		},
		{
			Name: "create with multiple inputs",
			Args: []string{processorName, cli.FunctionRefFlagName, functionRef, cli.InputFlagName, inputName, cli.InputFlagName, inputNameOther},
			ExpectCreates: []runtime.Object{
				&streamingv1alpha1.Processor{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      processorName,
					},
					Spec: streamingv1alpha1.ProcessorSpec{
						Build: &streamingv1alpha1.Build{FunctionRef: functionRef},
						Inputs: []streamingv1alpha1.InputStreamBinding{
							{Stream: inputName},
							{Stream: inputNameOther},
						},
						Template: &corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{},
								},
							},
						},
					},
				},
			},
			ExpectOutput: `
Created processor "my-processor"
`,
		},
		{
			Name: "create with single output",
			Args: []string{processorName, cli.FunctionRefFlagName, functionRef, cli.InputFlagName, inputName, cli.InputFlagName, inputNameOther, cli.OutputFlagName, outputName},
			ExpectCreates: []runtime.Object{
				&streamingv1alpha1.Processor{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      processorName,
					},
					Spec: streamingv1alpha1.ProcessorSpec{
						Build: &streamingv1alpha1.Build{FunctionRef: functionRef},
						Inputs: []streamingv1alpha1.InputStreamBinding{
							{Stream: inputName},
							{Stream: inputNameOther},
						},
						Outputs: []streamingv1alpha1.OutputStreamBinding{{Stream: outputName}},
						Template: &corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{},
								},
							},
						},
					},
				},
			},
			ExpectOutput: `
Created processor "my-processor"
`,
		},
		{
			Name: "create with some explicit parameter bindings",
			Args: []string{processorName, cli.FunctionRefFlagName, functionRef, cli.InputFlagName, inputNameBinding, cli.InputFlagName, inputNameOther, cli.OutputFlagName, outputNameOther, cli.OutputFlagName, outputNameBinding},
			ExpectCreates: []runtime.Object{
				&streamingv1alpha1.Processor{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      processorName,
					},
					Spec: streamingv1alpha1.ProcessorSpec{
						Build: &streamingv1alpha1.Build{FunctionRef: functionRef},
						Inputs: []streamingv1alpha1.InputStreamBinding{
							{Stream: inputName, Alias: inParameterName},
							{Stream: inputNameOther},
						},
						Outputs: []streamingv1alpha1.OutputStreamBinding{
							{Stream: outputNameOther},
							{Stream: outputName, Alias: outParameterName},
						},
						Template: &corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{},
								},
							},
						},
					},
				},
			},
			ExpectOutput: `
Created processor "my-processor"
`,
		},
		{
			Name: "create with multiple outputs",
			Args: []string{processorName, cli.FunctionRefFlagName, functionRef, cli.InputFlagName, inputName, cli.InputFlagName, inputNameOther, cli.OutputFlagName, outputName, cli.OutputFlagName, outputNameOther},
			ExpectCreates: []runtime.Object{
				&streamingv1alpha1.Processor{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      processorName,
					},
					Spec: streamingv1alpha1.ProcessorSpec{
						Build: &streamingv1alpha1.Build{FunctionRef: functionRef},
						Inputs: []streamingv1alpha1.InputStreamBinding{
							{Stream: inputName},
							{Stream: inputNameOther},
						},
						Outputs: []streamingv1alpha1.OutputStreamBinding{
							{Stream: outputName},
							{Stream: outputNameOther},
						},
						Template: &corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{},
								},
							},
						},
					},
				},
			},
			ExpectOutput: `
Created processor "my-processor"
`,
		},
		{
			Name: "error existing processor",
			Args: []string{processorName, cli.FunctionRefFlagName, functionRef, cli.InputFlagName, inputName},
			GivenObjects: []runtime.Object{
				&streamingv1alpha1.Processor{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      processorName,
					},
				},
			},
			ExpectCreates: []runtime.Object{
				&streamingv1alpha1.Processor{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      processorName,
					},
					Spec: streamingv1alpha1.ProcessorSpec{
						Build:  &streamingv1alpha1.Build{FunctionRef: functionRef},
						Inputs: []streamingv1alpha1.InputStreamBinding{{Stream: inputName}},
						Template: &corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{},
								},
							},
						},
					},
				},
			},
			ShouldError: true,
		},
		{
			Name: "error during create",
			Args: []string{processorName, cli.FunctionRefFlagName, functionRef, cli.InputFlagName, inputName},
			WithReactors: []rifftesting.ReactionFunc{
				rifftesting.InduceFailure("create", "processors"),
			},
			ExpectCreates: []runtime.Object{
				&streamingv1alpha1.Processor{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      processorName,
					},
					Spec: streamingv1alpha1.ProcessorSpec{
						Build:  &streamingv1alpha1.Build{FunctionRef: functionRef},
						Inputs: []streamingv1alpha1.InputStreamBinding{{Stream: inputName}},
						Template: &corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{},
								},
							},
						},
					},
				},
			},
			ShouldError: true,
		},
		{
			Name: "tail logs",
			Args: []string{processorName, cli.FunctionRefFlagName, functionRef, cli.InputFlagName, inputName, cli.TailFlagName},
			Prepare: func(t *testing.T, ctx context.Context, c *cli.Config) (context.Context, error) {
				lw := cachetesting.NewFakeControllerSource()
				ctx = k8s.WithListerWatcher(ctx, lw)

				kail := &kailtesting.Logger{}
				c.Kail = kail
				kail.On("StreamingProcessorLogs", mock.Anything, &streamingv1alpha1.Processor{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      processorName,
					},
					Spec: streamingv1alpha1.ProcessorSpec{
						Build:   &streamingv1alpha1.Build{FunctionRef: functionRef},
						Inputs:  []streamingv1alpha1.InputStreamBinding{{Stream: inputName}},
						Outputs: []streamingv1alpha1.OutputStreamBinding{},
						Template: &corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{},
								},
							},
						},
					},
				}, cli.TailSinceCreateDefault, mock.Anything).Return(nil).Run(func(args mock.Arguments) {
					fmt.Fprintf(c.Stdout, "...log output...\n")
				})
				return ctx, nil
			},
			CleanUp: func(t *testing.T, ctx context.Context, c *cli.Config) error {
				if lw, ok := k8s.GetListerWatcher(ctx, nil, "", nil).(*cachetesting.FakeControllerSource); ok {
					lw.Shutdown()
				}

				kail := c.Kail.(*kailtesting.Logger)
				kail.AssertExpectations(t)
				return nil
			},
			ExpectCreates: []runtime.Object{
				&streamingv1alpha1.Processor{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      processorName,
					},
					Spec: streamingv1alpha1.ProcessorSpec{
						Build:   &streamingv1alpha1.Build{FunctionRef: functionRef},
						Inputs:  []streamingv1alpha1.InputStreamBinding{{Stream: inputName}},
						Outputs: []streamingv1alpha1.OutputStreamBinding{},
						Template: &corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{},
								},
							},
						},
					},
				},
			},
			ExpectOutput: `
Created processor "my-processor"
Waiting for processor "my-processor" to become ready...
...log output...
Processor "my-processor" is ready
`,
		},
		{
			Name: "tail timeout",
			Args: []string{processorName, cli.FunctionRefFlagName, functionRef, cli.InputFlagName, inputName, cli.TailFlagName, cli.WaitTimeoutFlagName, "5ms"},
			Prepare: func(t *testing.T, ctx context.Context, c *cli.Config) (context.Context, error) {
				lw := cachetesting.NewFakeControllerSource()
				ctx = k8s.WithListerWatcher(ctx, lw)

				kail := &kailtesting.Logger{}
				c.Kail = kail
				kail.On("StreamingProcessorLogs", mock.Anything, &streamingv1alpha1.Processor{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      processorName,
					},
					Spec: streamingv1alpha1.ProcessorSpec{
						Build:   &streamingv1alpha1.Build{FunctionRef: functionRef},
						Inputs:  []streamingv1alpha1.InputStreamBinding{{Stream: inputName}},
						Outputs: []streamingv1alpha1.OutputStreamBinding{},
						Template: &corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{},
								},
							},
						},
					},
				}, cli.TailSinceCreateDefault, mock.Anything).Return(k8s.ErrWaitTimeout).Run(func(args mock.Arguments) {
					ctx := args[0].(context.Context)
					fmt.Fprintf(c.Stdout, "...log output...\n")
					// wait for context to be cancelled
					<-ctx.Done()
				})
				return ctx, nil
			},
			CleanUp: func(t *testing.T, ctx context.Context, c *cli.Config) error {
				if lw, ok := k8s.GetListerWatcher(ctx, nil, "", nil).(*cachetesting.FakeControllerSource); ok {
					lw.Shutdown()
				}

				kail := c.Kail.(*kailtesting.Logger)
				kail.AssertExpectations(t)
				return nil
			},
			ExpectCreates: []runtime.Object{
				&streamingv1alpha1.Processor{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      processorName,
					},
					Spec: streamingv1alpha1.ProcessorSpec{
						Build:   &streamingv1alpha1.Build{FunctionRef: functionRef},
						Inputs:  []streamingv1alpha1.InputStreamBinding{{Stream: inputName}},
						Outputs: []streamingv1alpha1.OutputStreamBinding{},
						Template: &corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{},
								},
							},
						},
					},
				},
			},
			ExpectOutput: `
Created processor "my-processor"
Waiting for processor "my-processor" to become ready...
...log output...
Timeout after "5ms" waiting for "my-processor" to become ready
To view status run: riff processor list --namespace default
To continue watching logs run: riff processor tail my-processor --namespace default
`,
			ShouldError: true,
			Verify: func(t *testing.T, output string, err error) {
				if actual := err; !errors.Is(err, cli.SilentError) {
					t.Errorf("expected error to be silent, actual %#v", actual)
				}
			},
		},
		{
			Name: "tail error",
			Args: []string{processorName, cli.FunctionRefFlagName, functionRef, cli.InputFlagName, inputName, cli.TailFlagName},
			Prepare: func(t *testing.T, ctx context.Context, c *cli.Config) (context.Context, error) {
				lw := cachetesting.NewFakeControllerSource()
				ctx = k8s.WithListerWatcher(ctx, lw)

				kail := &kailtesting.Logger{}
				c.Kail = kail
				kail.On("StreamingProcessorLogs", mock.Anything, &streamingv1alpha1.Processor{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      processorName,
					},
					Spec: streamingv1alpha1.ProcessorSpec{
						Build:   &streamingv1alpha1.Build{FunctionRef: functionRef},
						Inputs:  []streamingv1alpha1.InputStreamBinding{{Stream: inputName}},
						Outputs: []streamingv1alpha1.OutputStreamBinding{},
						Template: &corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{},
								},
							},
						},
					},
				}, cli.TailSinceCreateDefault, mock.Anything).Return(fmt.Errorf("kail error"))
				return ctx, nil
			},
			CleanUp: func(t *testing.T, ctx context.Context, c *cli.Config) error {
				if lw, ok := k8s.GetListerWatcher(ctx, nil, "", nil).(*cachetesting.FakeControllerSource); ok {
					lw.Shutdown()
				}

				kail := c.Kail.(*kailtesting.Logger)
				kail.AssertExpectations(t)
				return nil
			},
			ExpectCreates: []runtime.Object{
				&streamingv1alpha1.Processor{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      processorName,
					},
					Spec: streamingv1alpha1.ProcessorSpec{
						Build:   &streamingv1alpha1.Build{FunctionRef: functionRef},
						Inputs:  []streamingv1alpha1.InputStreamBinding{{Stream: inputName}},
						Outputs: []streamingv1alpha1.OutputStreamBinding{},
						Template: &corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{},
								},
							},
						},
					},
				},
			},
			ExpectOutput: `
Created processor "my-processor"
Waiting for processor "my-processor" to become ready...
`,
			ShouldError: true,
		},
		{
			Name: "create from function ref with env and env-from",
			Args: []string{processorName, cli.FunctionRefFlagName, functionRef, cli.InputFlagName, inputName, cli.EnvFlagName, envVar, cli.EnvFlagName, envVarOther, cli.EnvFromFlagName, envVarFromConfigMap, cli.EnvFromFlagName, envVarFromSecret},
			ExpectCreates: []runtime.Object{
				&streamingv1alpha1.Processor{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      processorName,
					},
					Spec: streamingv1alpha1.ProcessorSpec{
						Build:  &streamingv1alpha1.Build{FunctionRef: functionRef},
						Inputs: []streamingv1alpha1.InputStreamBinding{{Stream: inputName}},
						Template: &corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{
										Env: []corev1.EnvVar{
											{Name: envName, Value: envValue},
											{Name: envNameOther, Value: envValueOther},
											{
												Name: "MY_VAR_FROM_CONFIGMAP",
												ValueFrom: &corev1.EnvVarSource{
													ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
														LocalObjectReference: corev1.LocalObjectReference{
															Name: "my-configmap",
														},
														Key: "my-key",
													},
												},
											},
											{
												Name: "MY_VAR_FROM_SECRET",
												ValueFrom: &corev1.EnvVarSource{
													SecretKeyRef: &corev1.SecretKeySelector{
														LocalObjectReference: corev1.LocalObjectReference{
															Name: "my-secret",
														},
														Key: "my-key",
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
			ExpectOutput: `
Created processor "my-processor"
`,
		},
	}

	table.Run(t, commands.NewProcessorCreateCommand)
}
