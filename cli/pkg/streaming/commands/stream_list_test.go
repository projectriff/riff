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

	"github.com/projectriff/cli/pkg/cli"
	"github.com/projectriff/cli/pkg/streaming/commands"
	rifftesting "github.com/projectriff/cli/pkg/testing"
	"github.com/projectriff/system/pkg/apis"
	streamv1alpha1 "github.com/projectriff/system/pkg/apis/streaming/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestStreamListOptions(t *testing.T) {
	table := rifftesting.OptionsTable{
		{
			Name: "invalid list",
			Options: &commands.StreamListOptions{
				ListOptions: rifftesting.InvalidListOptions,
			},
			ExpectFieldErrors: rifftesting.InvalidListOptionsFieldError,
		},
		{
			Name: "valid list",
			Options: &commands.StreamListOptions{
				ListOptions: rifftesting.ValidListOptions,
			},
			ShouldValidate: true,
		},
	}

	table.Run(t)
}

func TestStreamListCommand(t *testing.T) {
	streamName := "test-stream"
	streamOtherName := "test-other-stream"
	defaultNamespace := "default"
	otherNamespace := "other-namespace"

	table := rifftesting.CommandTable{
		{
			Name: "invalid args",
			Args: []string{},
			Prepare: func(t *testing.T, ctx context.Context, c *cli.Config) (context.Context, error) {
				// disable default namespace
				c.Client.(*rifftesting.FakeClient).Namespace = ""
				return ctx, nil
			},
			ShouldError: true,
		},
		{
			Name: "empty",
			Args: []string{},
			ExpectOutput: `
No streams found.
`,
		},
		{
			Name: "lists an item",
			Args: []string{},
			GivenObjects: []runtime.Object{
				&streamv1alpha1.Stream{
					ObjectMeta: metav1.ObjectMeta{
						Name:      streamName,
						Namespace: defaultNamespace,
					},
				},
			},
			ExpectOutput: `
NAME          GATEWAY   CONTENT-TYPE   STATUS      AGE
test-stream   <empty>   <empty>        <unknown>   <unknown>
`,
		},
		{
			Name: "filters by namespace",
			Args: []string{cli.NamespaceFlagName, otherNamespace},
			GivenObjects: []runtime.Object{
				&streamv1alpha1.Stream{
					ObjectMeta: metav1.ObjectMeta{
						Name:      streamName,
						Namespace: defaultNamespace,
					},
				},
			},
			ExpectOutput: `
No streams found.
`,
		},
		{
			Name: "all namespace",
			Args: []string{cli.AllNamespacesFlagName},
			GivenObjects: []runtime.Object{
				&streamv1alpha1.Stream{
					ObjectMeta: metav1.ObjectMeta{
						Name:      streamName,
						Namespace: defaultNamespace,
					},
				},
				&streamv1alpha1.Stream{
					ObjectMeta: metav1.ObjectMeta{
						Name:      streamOtherName,
						Namespace: otherNamespace,
					},
				},
			},
			ExpectOutput: `
NAMESPACE         NAME                GATEWAY   CONTENT-TYPE   STATUS      AGE
default           test-stream         <empty>   <empty>        <unknown>   <unknown>
other-namespace   test-other-stream   <empty>   <empty>        <unknown>   <unknown>
`,
		},
		{
			Name: "table populates all columns",
			Args: []string{},
			GivenObjects: []runtime.Object{
				&streamv1alpha1.Stream{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "words",
						Namespace: defaultNamespace,
					},
					Spec: streamv1alpha1.StreamSpec{
						Gateway:     corev1.LocalObjectReference{Name: "kafka"},
						ContentType: "text/csv",
					},
					Status: streamv1alpha1.StreamStatus{
						Status: apis.Status{
							Conditions: apis.Conditions{
								{Type: streamv1alpha1.StreamConditionReady, Status: "True"},
							},
						},
					},
				},
			},
			ExpectOutput: `
NAME    GATEWAY   CONTENT-TYPE   STATUS   AGE
words   kafka     text/csv       Ready    <unknown>
`,
		},
		{
			Name: "list error",
			Args: []string{},
			WithReactors: []rifftesting.ReactionFunc{
				rifftesting.InduceFailure("list", "streams"),
			},
			ShouldError: true,
		},
	}

	table.Run(t, commands.NewStreamListCommand)
}
