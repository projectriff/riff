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
	"github.com/projectriff/cli/pkg/core/commands"
	"github.com/projectriff/cli/pkg/k8s"
	rifftesting "github.com/projectriff/cli/pkg/testing"
	kailtesting "github.com/projectriff/cli/pkg/testing/kail"
	corev1alpha1 "github.com/projectriff/system/pkg/apis/core/v1alpha1"
	"github.com/stretchr/testify/mock"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	cachetesting "k8s.io/client-go/tools/cache/testing"
)

func TestDeployerCreateOptions(t *testing.T) {
	table := rifftesting.OptionsTable{
		{
			Name: "invalid resource",
			Options: &commands.DeployerCreateOptions{
				ResourceOptions: rifftesting.InvalidResourceOptions,
			},
			ExpectFieldErrors: rifftesting.InvalidResourceOptionsFieldError.Also(
				cli.ErrMissingOneOf(cli.ApplicationRefFlagName, cli.ContainerRefFlagName, cli.FunctionRefFlagName, cli.ImageFlagName),
				cli.ErrInvalidValue("", cli.IngressPolicyFlagName),
			),
		},
		{
			Name: "from application",
			Options: &commands.DeployerCreateOptions{
				ResourceOptions: rifftesting.ValidResourceOptions,
				ApplicationRef:  "my-application",
				IngressPolicy:   string(corev1alpha1.IngressPolicyClusterLocal),
			},
			ShouldValidate: true,
		},
		{
			Name: "from container",
			Options: &commands.DeployerCreateOptions{
				ResourceOptions: rifftesting.ValidResourceOptions,
				ContainerRef:    "my-container",
				IngressPolicy:   string(corev1alpha1.IngressPolicyClusterLocal),
			},
			ShouldValidate: true,
		},
		{
			Name: "from function",
			Options: &commands.DeployerCreateOptions{
				ResourceOptions: rifftesting.ValidResourceOptions,
				FunctionRef:     "my-function",
				IngressPolicy:   string(corev1alpha1.IngressPolicyClusterLocal),
			},
			ShouldValidate: true,
		},
		{
			Name: "from image",
			Options: &commands.DeployerCreateOptions{
				ResourceOptions: rifftesting.ValidResourceOptions,
				Image:           "example.com/repo:tag",
				IngressPolicy:   string(corev1alpha1.IngressPolicyClusterLocal),
			},
			ShouldValidate: true,
		},
		{
			Name: "from application, container, function and image",
			Options: &commands.DeployerCreateOptions{
				ResourceOptions: rifftesting.ValidResourceOptions,
				ApplicationRef:  "my-application",
				ContainerRef:    "my-container",
				FunctionRef:     "my-function",
				Image:           "example.com/repo:tag",
				IngressPolicy:   string(corev1alpha1.IngressPolicyClusterLocal),
			},
			ExpectFieldErrors: cli.ErrMultipleOneOf(cli.ApplicationRefFlagName, cli.ContainerRefFlagName, cli.FunctionRefFlagName, cli.ImageFlagName),
		},
		{
			Name: "with external ingress",
			Options: &commands.DeployerCreateOptions{
				ResourceOptions: rifftesting.ValidResourceOptions,
				Image:           "example.com/repo:tag",
				IngressPolicy:   string(corev1alpha1.IngressPolicyExternal),
			},
			ShouldValidate: true,
		},
		{
			Name: "with bogus ingress",
			Options: &commands.DeployerCreateOptions{
				ResourceOptions: rifftesting.ValidResourceOptions,
				Image:           "example.com/repo:tag",
				IngressPolicy:   "bogus",
			},
			ExpectFieldErrors: cli.ErrInvalidValue("bogus", cli.IngressPolicyFlagName),
		},
		{
			Name: "with env",
			Options: &commands.DeployerCreateOptions{
				ResourceOptions: rifftesting.ValidResourceOptions,
				Image:           "example.com/repo:tag",
				IngressPolicy:   string(corev1alpha1.IngressPolicyClusterLocal),
				Env:             []string{"VAR1=foo", "VAR2=bar"},
			},
			ShouldValidate: true,
		},
		{
			Name: "with invalid env",
			Options: &commands.DeployerCreateOptions{
				ResourceOptions: rifftesting.ValidResourceOptions,
				Image:           "example.com/repo:tag",
				IngressPolicy:   string(corev1alpha1.IngressPolicyClusterLocal),
				Env:             []string{"=foo"},
			},
			ExpectFieldErrors: cli.ErrInvalidArrayValue("=foo", cli.EnvFlagName, 0),
		},
		{
			Name: "with envfrom secret",
			Options: &commands.DeployerCreateOptions{
				ResourceOptions: rifftesting.ValidResourceOptions,
				Image:           "example.com/repo:tag",
				IngressPolicy:   string(corev1alpha1.IngressPolicyClusterLocal),
				EnvFrom:         []string{"VAR1=secretKeyRef:name:key"},
			},
			ShouldValidate: true,
		},
		{
			Name: "with envfrom configmap",
			Options: &commands.DeployerCreateOptions{
				ResourceOptions: rifftesting.ValidResourceOptions,
				Image:           "example.com/repo:tag",
				IngressPolicy:   string(corev1alpha1.IngressPolicyClusterLocal),
				EnvFrom:         []string{"VAR1=configMapKeyRef:name:key"},
			},
			ShouldValidate: true,
		},
		{
			Name: "with invalid envfrom",
			Options: &commands.DeployerCreateOptions{
				ResourceOptions: rifftesting.ValidResourceOptions,
				Image:           "example.com/repo:tag",
				IngressPolicy:   string(corev1alpha1.IngressPolicyClusterLocal),
				EnvFrom:         []string{"VAR1=someOtherKeyRef:name:key"},
			},
			ExpectFieldErrors: cli.ErrInvalidArrayValue("VAR1=someOtherKeyRef:name:key", cli.EnvFromFlagName, 0),
		},
		{
			Name: "with limits",
			Options: &commands.DeployerCreateOptions{
				ResourceOptions: rifftesting.ValidResourceOptions,
				Image:           "example.com/repo:tag",
				IngressPolicy:   string(corev1alpha1.IngressPolicyClusterLocal),
				LimitCPU:        "500m",
				LimitMemory:     "512Mi",
			},
			ShouldValidate: true,
		},
		{
			Name: "with invalid limits",
			Options: &commands.DeployerCreateOptions{
				ResourceOptions: rifftesting.ValidResourceOptions,
				Image:           "example.com/repo:tag",
				IngressPolicy:   string(corev1alpha1.IngressPolicyClusterLocal),
				LimitCPU:        "50%",
				LimitMemory:     "NaN",
			},
			ExpectFieldErrors: cli.FieldErrors{}.Also(
				cli.ErrInvalidValue("50%", cli.LimitCPUFlagName),
				cli.ErrInvalidValue("NaN", cli.LimitMemoryFlagName),
			),
		},
		{
			Name: "with target-port",
			Options: &commands.DeployerCreateOptions{
				ResourceOptions: rifftesting.ValidResourceOptions,
				Image:           "example.com/repo:tag",
				IngressPolicy:   string(corev1alpha1.IngressPolicyClusterLocal),
				TargetPort:      8888,
			},
			ShouldValidate: true,
		},
		{
			Name: "with invalid target-port",
			Options: &commands.DeployerCreateOptions{
				ResourceOptions: rifftesting.ValidResourceOptions,
				Image:           "example.com/repo:tag",
				IngressPolicy:   string(corev1alpha1.IngressPolicyClusterLocal),
				TargetPort:      -1,
			},
			ExpectFieldErrors: cli.FieldErrors{}.Also(
				cli.ErrInvalidValue("-1", cli.TargetPortFlagName),
			),
		},
		{
			Name: "with tail",
			Options: &commands.DeployerCreateOptions{
				ResourceOptions: rifftesting.ValidResourceOptions,
				Image:           "example.com/repo:tag",
				IngressPolicy:   string(corev1alpha1.IngressPolicyClusterLocal),
				Tail:            true,
				WaitTimeout:     "10m",
			},
			ShouldValidate: true,
		},
		{
			Name: "with tail, missing timeout",
			Options: &commands.DeployerCreateOptions{
				ResourceOptions: rifftesting.ValidResourceOptions,
				Image:           "example.com/repo:tag",
				IngressPolicy:   string(corev1alpha1.IngressPolicyClusterLocal),
				Tail:            true,
			},
			ExpectFieldErrors: cli.ErrMissingField(cli.WaitTimeoutFlagName),
		},
		{
			Name: "with tail, invalid timeout",
			Options: &commands.DeployerCreateOptions{
				ResourceOptions: rifftesting.ValidResourceOptions,
				Image:           "example.com/repo:tag",
				IngressPolicy:   string(corev1alpha1.IngressPolicyClusterLocal),
				Tail:            true,
				WaitTimeout:     "d",
			},
			ExpectFieldErrors: cli.ErrInvalidValue("d", cli.WaitTimeoutFlagName),
		},
		{
			Name: "dry run",
			Options: &commands.DeployerCreateOptions{
				ResourceOptions: rifftesting.ValidResourceOptions,
				Image:           "example.com/repo:tag",
				IngressPolicy:   string(corev1alpha1.IngressPolicyClusterLocal),
				DryRun:          true,
			},
			ShouldValidate: true,
		},
		{
			Name: "dry run, tail",
			Options: &commands.DeployerCreateOptions{
				ResourceOptions: rifftesting.ValidResourceOptions,
				Image:           "example.com/repo:tag",
				IngressPolicy:   string(corev1alpha1.IngressPolicyClusterLocal),
				Tail:            true,
				WaitTimeout:     "10m",
				DryRun:          true,
			},
			ExpectFieldErrors: cli.ErrMultipleOneOf(cli.DryRunFlagName, cli.TailFlagName),
		},
	}

	table.Run(t)
}

func TestDeployerCreateCommand(t *testing.T) {
	defaultNamespace := "default"
	deployerName := "my-deployer"
	image := "registry.example.com/repo@sha256:deadbeefdeadbeefdeadbeefdeadbeef"
	applicationRef := "my-app"
	containerRef := "my-container"
	containerPort := int32(8888)
	functionRef := "my-func"
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
			Name: "create from image",
			Args: []string{deployerName, cli.ImageFlagName, image},
			ExpectCreates: []runtime.Object{
				&corev1alpha1.Deployer{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      deployerName,
					},
					Spec: corev1alpha1.DeployerSpec{
						Template: &corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{Image: image},
								},
							},
						},
						IngressPolicy: corev1alpha1.IngressPolicyClusterLocal,
					},
				},
			},
			ExpectOutput: `
Created deployer "my-deployer"
`,
		},
		{
			Name: "create from application ref",
			Args: []string{deployerName, cli.ApplicationRefFlagName, applicationRef},
			ExpectCreates: []runtime.Object{
				&corev1alpha1.Deployer{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      deployerName,
					},
					Spec: corev1alpha1.DeployerSpec{
						Build: &corev1alpha1.Build{
							ApplicationRef: applicationRef,
						},
						IngressPolicy: corev1alpha1.IngressPolicyClusterLocal,
					},
				},
			},
			ExpectOutput: `
Created deployer "my-deployer"
`,
		},
		{
			Name: "create from container ref",
			Args: []string{deployerName, cli.ContainerRefFlagName, containerRef},
			ExpectCreates: []runtime.Object{
				&corev1alpha1.Deployer{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      deployerName,
					},
					Spec: corev1alpha1.DeployerSpec{
						Build: &corev1alpha1.Build{
							ContainerRef: containerRef,
						},
						IngressPolicy: corev1alpha1.IngressPolicyClusterLocal,
					},
				},
			},
			ExpectOutput: `
Created deployer "my-deployer"
`,
		},
		{
			Name: "create from function ref",
			Args: []string{deployerName, cli.FunctionRefFlagName, functionRef},
			ExpectCreates: []runtime.Object{
				&corev1alpha1.Deployer{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      deployerName,
					},
					Spec: corev1alpha1.DeployerSpec{
						Build: &corev1alpha1.Build{
							FunctionRef: functionRef,
						},
						IngressPolicy: corev1alpha1.IngressPolicyClusterLocal,
					},
				},
			},
			ExpectOutput: `
Created deployer "my-deployer"
`,
		},
		{
			Name: "dry run",
			Args: []string{deployerName, cli.ImageFlagName, image, cli.DryRunFlagName},
			ExpectOutput: `
---
apiVersion: core.projectriff.io/v1alpha1
kind: Deployer
metadata:
  creationTimestamp: null
  name: my-deployer
  namespace: default
spec:
  ingressPolicy: ClusterLocal
  template:
    metadata:
      creationTimestamp: null
    spec:
      containers:
      - image: registry.example.com/repo@sha256:deadbeefdeadbeefdeadbeefdeadbeef
        name: ""
        resources: {}
status: {}

Created deployer "my-deployer"
`,
		},
		{
			Name: "create from external ingress policy",
			Args: []string{deployerName, cli.ImageFlagName, image, cli.IngressPolicyFlagName, string(corev1alpha1.IngressPolicyExternal)},
			ExpectCreates: []runtime.Object{
				&corev1alpha1.Deployer{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      deployerName,
					},
					Spec: corev1alpha1.DeployerSpec{
						Template: &corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{Image: image},
								},
							},
						},
						IngressPolicy: corev1alpha1.IngressPolicyExternal,
					},
				},
			},
			ExpectOutput: `
Created deployer "my-deployer"
`,
		},
		{
			Name: "create from image with env and env-from",
			Args: []string{deployerName, cli.ImageFlagName, image, cli.EnvFlagName, envVar, cli.EnvFlagName, envVarOther, cli.EnvFromFlagName, envVarFromConfigMap, cli.EnvFromFlagName, envVarFromSecret},
			ExpectCreates: []runtime.Object{
				&corev1alpha1.Deployer{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      deployerName,
					},
					Spec: corev1alpha1.DeployerSpec{
						Template: &corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{
										Image: image,
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
						IngressPolicy: corev1alpha1.IngressPolicyClusterLocal,
					},
				},
			},
			ExpectOutput: `
Created deployer "my-deployer"
`,
		},
		{
			Name: "create with limits",
			Args: []string{deployerName, cli.ImageFlagName, image, cli.LimitCPUFlagName, "100m", cli.LimitMemoryFlagName, "128Mi"},
			ExpectCreates: []runtime.Object{
				&corev1alpha1.Deployer{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      deployerName,
					},
					Spec: corev1alpha1.DeployerSpec{
						Template: &corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{
										Image: image,
										Resources: corev1.ResourceRequirements{
											Limits: corev1.ResourceList{
												corev1.ResourceCPU:    resource.MustParse("100m"),
												corev1.ResourceMemory: resource.MustParse("128Mi"),
											},
										},
									},
								},
							},
						},
						IngressPolicy: corev1alpha1.IngressPolicyClusterLocal,
					},
				},
			},
			ExpectOutput: `
Created deployer "my-deployer"
`,
		},
		{
			Name: "create with target-port",
			Args: []string{deployerName, cli.ImageFlagName, image, cli.TargetPortFlagName, "8888"},
			ExpectCreates: []runtime.Object{
				&corev1alpha1.Deployer{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      deployerName,
					},
					Spec: corev1alpha1.DeployerSpec{
						Template: &corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{
										Image: image,
										Ports: []corev1.ContainerPort{
											{Protocol: corev1.ProtocolTCP, ContainerPort: containerPort},
										},
									},
								},
							},
						},
						IngressPolicy: corev1alpha1.IngressPolicyClusterLocal,
					},
				},
			},
			ExpectOutput: `
Created deployer "my-deployer"
`,
		},
		{
			Name: "error existing deployer",
			Args: []string{deployerName, cli.ImageFlagName, image},
			GivenObjects: []runtime.Object{
				&corev1alpha1.Deployer{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      deployerName,
					},
				},
			},
			ExpectCreates: []runtime.Object{
				&corev1alpha1.Deployer{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      deployerName,
					},
					Spec: corev1alpha1.DeployerSpec{
						Template: &corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{Image: image},
								},
							},
						},
						IngressPolicy: corev1alpha1.IngressPolicyClusterLocal,
					},
				},
			},
			ShouldError: true,
		},
		{
			Name: "error during create",
			Args: []string{deployerName, cli.ImageFlagName, image},
			WithReactors: []rifftesting.ReactionFunc{
				rifftesting.InduceFailure("create", "deployers"),
			},
			ExpectCreates: []runtime.Object{
				&corev1alpha1.Deployer{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      deployerName,
					},
					Spec: corev1alpha1.DeployerSpec{
						Template: &corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{Image: image},
								},
							},
						},
						IngressPolicy: corev1alpha1.IngressPolicyClusterLocal,
					},
				},
			},
			ShouldError: true,
		},
		{
			Name: "tail logs",
			Args: []string{deployerName, cli.ImageFlagName, image, cli.TailFlagName},
			Prepare: func(t *testing.T, ctx context.Context, c *cli.Config) (context.Context, error) {
				lw := cachetesting.NewFakeControllerSource()
				ctx = k8s.WithListerWatcher(ctx, lw)

				kail := &kailtesting.Logger{}
				c.Kail = kail
				kail.On("CoreDeployerLogs", mock.Anything, &corev1alpha1.Deployer{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      deployerName,
					},
					Spec: corev1alpha1.DeployerSpec{
						Template: &corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{{Image: image}},
							},
						},
						IngressPolicy: corev1alpha1.IngressPolicyClusterLocal,
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
				&corev1alpha1.Deployer{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      deployerName,
					},
					Spec: corev1alpha1.DeployerSpec{
						Template: &corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{{Image: image}},
							},
						},
						IngressPolicy: corev1alpha1.IngressPolicyClusterLocal,
					},
				},
			},
			ExpectOutput: `
Created deployer "my-deployer"
Waiting for deployer "my-deployer" to become ready...
...log output...
Deployer "my-deployer" is ready
`,
		},
		{
			Name: "tail timeout",
			Args: []string{deployerName, cli.ImageFlagName, image, cli.TailFlagName, cli.WaitTimeoutFlagName, "5ms"},
			Prepare: func(t *testing.T, ctx context.Context, c *cli.Config) (context.Context, error) {
				lw := cachetesting.NewFakeControllerSource()
				ctx = k8s.WithListerWatcher(ctx, lw)

				kail := &kailtesting.Logger{}
				c.Kail = kail
				kail.On("CoreDeployerLogs", mock.Anything, &corev1alpha1.Deployer{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      deployerName,
					},
					Spec: corev1alpha1.DeployerSpec{
						Template: &corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{{Image: image}},
							},
						},
						IngressPolicy: corev1alpha1.IngressPolicyClusterLocal,
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
				&corev1alpha1.Deployer{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      deployerName,
					},
					Spec: corev1alpha1.DeployerSpec{
						Template: &corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{{Image: image}},
							},
						},
						IngressPolicy: corev1alpha1.IngressPolicyClusterLocal,
					},
				},
			},
			ExpectOutput: `
Created deployer "my-deployer"
Waiting for deployer "my-deployer" to become ready...
...log output...
Timeout after "5ms" waiting for "my-deployer" to become ready
To view status run: riff core deployer list --namespace default
To continue watching logs run: riff core deployer tail my-deployer --namespace default
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
			Args: []string{deployerName, cli.ImageFlagName, image, cli.TailFlagName},
			Prepare: func(t *testing.T, ctx context.Context, c *cli.Config) (context.Context, error) {
				lw := cachetesting.NewFakeControllerSource()
				ctx = k8s.WithListerWatcher(ctx, lw)

				kail := &kailtesting.Logger{}
				c.Kail = kail
				kail.On("CoreDeployerLogs", mock.Anything, &corev1alpha1.Deployer{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      deployerName,
					},
					Spec: corev1alpha1.DeployerSpec{
						Template: &corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{{Image: image}},
							},
						},
						IngressPolicy: corev1alpha1.IngressPolicyClusterLocal,
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
				&corev1alpha1.Deployer{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      deployerName,
					},
					Spec: corev1alpha1.DeployerSpec{
						Template: &corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{{Image: image}},
							},
						},
						IngressPolicy: corev1alpha1.IngressPolicyClusterLocal,
					},
				},
			},
			ExpectOutput: `
Created deployer "my-deployer"
Waiting for deployer "my-deployer" to become ready...
`,
			ShouldError: true,
		},
	}

	table.Run(t, commands.NewDeployerCreateCommand)
}
