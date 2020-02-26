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
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	clientgotesting "k8s.io/client-go/testing"
	cachetesting "k8s.io/client-go/tools/cache/testing"

	"github.com/projectriff/cli/pkg/cli"
	"github.com/projectriff/cli/pkg/k8s"
	"github.com/projectriff/cli/pkg/streaming/commands"
	rifftesting "github.com/projectriff/cli/pkg/testing"
	streamv1alpha1 "github.com/projectriff/system/pkg/apis/streaming/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestStreamCreateOptions(t *testing.T) {
	table := rifftesting.OptionsTable{
		{
			Name: "invalid resource",
			Options: &commands.StreamCreateOptions{
				ResourceOptions: rifftesting.InvalidResourceOptions,
			},
			ExpectFieldErrors: rifftesting.InvalidResourceOptionsFieldError.Also(
				cli.ErrMissingField(cli.GatewayFlagName),
			),
		},
		{
			Name: "valid gateway",
			Options: &commands.StreamCreateOptions{
				ResourceOptions: rifftesting.ValidResourceOptions,
				Gateway:         "test-gateway",
			},
			ShouldValidate: true,
		},
		{
			Name: "no gateway",
			Options: &commands.StreamCreateOptions{
				ResourceOptions: rifftesting.ValidResourceOptions,
			},
			ExpectFieldErrors: cli.ErrMissingField(cli.GatewayFlagName),
		},
		{
			Name: "with valid content type",
			Options: &commands.StreamCreateOptions{
				ResourceOptions: rifftesting.ValidResourceOptions,
				Gateway:         "test-gateway",
				ContentType:     "application/x-doom",
			},
			ShouldValidate: true,
		},
		{
			Name: "with invalid content-type",
			Options: &commands.StreamCreateOptions{
				ResourceOptions: rifftesting.ValidResourceOptions,
				Gateway:         "test-gateway",
				ContentType:     "invalid-content-type",
			},
			ExpectFieldErrors: cli.ErrInvalidValue("invalid-content-type", cli.ContentTypeFlagName),
		},
		{
			Name: "dry run",
			Options: &commands.StreamCreateOptions{
				ResourceOptions: rifftesting.ValidResourceOptions,
				Gateway:         "test-gateway",
				DryRun:          true,
			},
			ShouldValidate: true,
		},
		{
			Name: "dry run, tail",
			Options: &commands.StreamCreateOptions{
				ResourceOptions: rifftesting.ValidResourceOptions,
				Gateway:         "test-gateway",
				DryRun:          true,
				Tail:            true,
			},
			ExpectFieldErrors: cli.ErrMultipleOneOf(cli.DryRunFlagName, cli.TailFlagName),
		},
		{
			Name: "negative timeout",
			Options: &commands.StreamCreateOptions{
				ResourceOptions: rifftesting.ValidResourceOptions,
				Gateway:         "test-gateway",
				WaitTimeout:     -3 * time.Second,
			},
			ExpectFieldErrors: cli.ErrInvalidValue(-3*time.Second, cli.WaitTimeoutFlagName),
		},
	}

	table.Run(t)
}

func TestStreamCreateCommand(t *testing.T) {
	defaultNamespace := "default"
	streamName := "my-stream"
	defaultContentType := "application/octet-stream"
	contentType := "video/jpeg"
	gateway := "test-gateway"

	var lister *cachetesting.FakeControllerSource

	table := rifftesting.CommandTable{
		{
			Name:        "invalid args",
			Args:        []string{},
			ShouldError: true,
		},
		{
			Name: "stream gateway",
			Args: []string{streamName, cli.GatewayFlagName, gateway},
			ExpectCreates: []runtime.Object{
				&streamv1alpha1.Stream{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      streamName,
					},
					Spec: streamv1alpha1.StreamSpec{
						Gateway:     corev1.LocalObjectReference{Name: gateway},
						ContentType: defaultContentType,
					},
				},
			},
			ExpectOutput: `
Created stream "my-stream"
`,
		},
		{
			Name: "dry run",
			Args: []string{streamName, cli.GatewayFlagName, gateway, cli.DryRunFlagName},
			ExpectOutput: `
---
apiVersion: streaming.projectriff.io/v1alpha1
kind: Stream
metadata:
  creationTimestamp: null
  name: my-stream
  namespace: default
spec:
  contentType: ""
  gateway:
    name: test-gateway
status:
  binding:
    metadataRef: {}
    secretRef: {}

Created stream "my-stream"
`,
		},
		{
			Name: "with optional content-type",
			Args: []string{streamName, cli.GatewayFlagName, gateway, cli.ContentTypeFlagName, contentType},
			ExpectCreates: []runtime.Object{
				&streamv1alpha1.Stream{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      streamName,
					},
					Spec: streamv1alpha1.StreamSpec{
						Gateway:     corev1.LocalObjectReference{Name: gateway},
						ContentType: contentType,
					},
				},
			},
			ExpectOutput: `
Created stream "my-stream"
`,
		},
		{
			Name: "error existing stream",
			Args: []string{streamName, cli.GatewayFlagName, gateway},
			GivenObjects: []runtime.Object{
				&streamv1alpha1.Stream{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      streamName,
					},
				},
			},
			ExpectCreates: []runtime.Object{
				&streamv1alpha1.Stream{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      streamName,
					},
					Spec: streamv1alpha1.StreamSpec{
						Gateway: corev1.LocalObjectReference{Name: gateway},
					},
				},
			},
			ShouldError: true,
		},
		{
			Name: "error during create",
			Args: []string{streamName, cli.GatewayFlagName, gateway},
			WithReactors: []rifftesting.ReactionFunc{
				rifftesting.InduceFailure("create", "streams"),
			},
			ExpectCreates: []runtime.Object{
				&streamv1alpha1.Stream{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      streamName,
					},
					Spec: streamv1alpha1.StreamSpec{
						Gateway: corev1.LocalObjectReference{Name: gateway},
					},
				},
			},
			ShouldError: true,
		},
		{
			Name: "tail",
			Args: []string{"input", cli.GatewayFlagName, "franz", cli.TailFlagName, cli.ContentTypeFlagName, "application/json"},
			Prepare: func(t *testing.T, ctx context.Context, c *cli.Config) (context.Context, error) {
				lister = cachetesting.NewFakeControllerSource()
				ctx = k8s.WithListerWatcher(ctx, lister)

				return ctx, nil
			},
			WithReactors: []rifftesting.ReactionFunc{
				func(action clientgotesting.Action) (handled bool, ret runtime.Object, err error) {
					if c, ok := action.(clientgotesting.CreateAction); ok {
						copy := c.GetObject().DeepCopyObject()
						t := time.NewTimer(time.Millisecond * 200)
						go func() {
							<-t.C
							copy.(*streamv1alpha1.Stream).Status.MarkBindingReady()
							copy.(*streamv1alpha1.Stream).Status.MarkStreamProvisioned()
							lister.Modify(copy)
						}()
					}
					return false, nil, nil
				},
			},
			CleanUp: func(t *testing.T, ctx context.Context, c *cli.Config) error {
				if lw, ok := k8s.GetListerWatcher(ctx, nil, "", nil).(*cachetesting.FakeControllerSource); ok {
					lw.Shutdown()
				}
				lister = nil
				return nil
			},
			ExpectCreates: []runtime.Object{
				&streamv1alpha1.Stream{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      "input",
					},
					Spec: streamv1alpha1.StreamSpec{
						ContentType: "application/json",
						Gateway:     corev1.LocalObjectReference{Name: "franz"},
					},
				},
			},
			ExpectOutput: `
Created stream "input"
Waiting for stream "input" to become ready...
Stream "input" is ready
`,
		},
		{
			Name: "tail timeout",
			Args: []string{"input", cli.GatewayFlagName, "franz", cli.TailFlagName, cli.ContentTypeFlagName, "application/json", cli.WaitTimeoutFlagName, "10ms"},
			Prepare: func(t *testing.T, ctx context.Context, c *cli.Config) (context.Context, error) {
				lister = cachetesting.NewFakeControllerSource()
				ctx = k8s.WithListerWatcher(ctx, lister)

				return ctx, nil
			},
			CleanUp: func(t *testing.T, ctx context.Context, c *cli.Config) error {
				if lw, ok := k8s.GetListerWatcher(ctx, nil, "", nil).(*cachetesting.FakeControllerSource); ok {
					lw.Shutdown()
				}
				lister = nil
				return nil
			},
			ExpectCreates: []runtime.Object{
				&streamv1alpha1.Stream{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      "input",
					},
					Spec: streamv1alpha1.StreamSpec{
						ContentType: "application/json",
						Gateway:     corev1.LocalObjectReference{Name: "franz"},
					},
				},
			},
			ShouldError: true,
			ExpectOutput: `
Created stream "input"
Waiting for stream "input" to become ready...
Timeout after "10ms" waiting for "input" to become ready
To view status run: riff streaming stream list --namespace default
`,
		},
	}

	table.Run(t, commands.NewStreamCreateCommand)
}
