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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestProcessorListOptions(t *testing.T) {
	table := rifftesting.OptionsTable{
		{
			Name: "invalid list",
			Options: &commands.ProcessorListOptions{
				ListOptions: rifftesting.InvalidListOptions,
			},
			ExpectFieldErrors: rifftesting.InvalidListOptionsFieldError,
		},
		{
			Name: "valid list",
			Options: &commands.ProcessorListOptions{
				ListOptions: rifftesting.ValidListOptions,
			},
			ShouldValidate: true,
		},
	}

	table.Run(t)
}

func TestProcessorListCommand(t *testing.T) {
	processorName := "test-processor"
	processorOtherName := "test-other-processor"
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
No processors found.
`,
		},
		{
			Name: "lists an item",
			Args: []string{},
			GivenObjects: []runtime.Object{
				&streamv1alpha1.Processor{
					ObjectMeta: metav1.ObjectMeta{
						Name:      processorName,
						Namespace: defaultNamespace,
					},
				},
			},
			ExpectOutput: `
NAME             FUNCTION   INPUTS    OUTPUTS   STATUS      AGE
test-processor   <empty>    <empty>   <empty>   <unknown>   <unknown>
`,
		},
		{
			Name: "filters by namespace",
			Args: []string{cli.NamespaceFlagName, otherNamespace},
			GivenObjects: []runtime.Object{
				&streamv1alpha1.Processor{
					ObjectMeta: metav1.ObjectMeta{
						Name:      processorName,
						Namespace: defaultNamespace,
					},
				},
			},
			ExpectOutput: `
No processors found.
`,
		},
		{
			Name: "all namespace",
			Args: []string{cli.AllNamespacesFlagName},
			GivenObjects: []runtime.Object{
				&streamv1alpha1.Processor{
					ObjectMeta: metav1.ObjectMeta{
						Name:      processorName,
						Namespace: defaultNamespace,
					},
				},
				&streamv1alpha1.Processor{
					ObjectMeta: metav1.ObjectMeta{
						Name:      processorOtherName,
						Namespace: otherNamespace,
					},
				},
			},
			ExpectOutput: `
NAMESPACE         NAME                   FUNCTION   INPUTS    OUTPUTS   STATUS      AGE
default           test-processor         <empty>    <empty>   <empty>   <unknown>   <unknown>
other-namespace   test-other-processor   <empty>    <empty>   <empty>   <unknown>   <unknown>
`,
		},
		{
			Name: "table populates all columns",
			Args: []string{},
			GivenObjects: []runtime.Object{
				&streamv1alpha1.Processor{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "square",
						Namespace: defaultNamespace,
					},
					Spec: streamv1alpha1.ProcessorSpec{
						Build:   &streamv1alpha1.Build{FunctionRef: "square"},
						Inputs:  []streamv1alpha1.InputStreamBinding{{Stream: "numbers", Alias: "n1"}, {Stream: "morenumbers", Alias: "n2"}},
						Outputs: []streamv1alpha1.OutputStreamBinding{{Stream: "squares", Alias: "s"}},
					},
					Status: streamv1alpha1.ProcessorStatus{
						Status: apis.Status{
							Conditions: apis.Conditions{
								{Type: streamv1alpha1.ProcessorConditionReady, Status: "True"},
							},
						},
					},
				},
			},
			ExpectOutput: `
NAME     FUNCTION   INPUTS                       OUTPUTS     STATUS   AGE
square   square     n1:numbers, n2:morenumbers   s:squares   Ready    <unknown>
`,
		},
		{
			Name: "list error",
			Args: []string{},
			WithReactors: []rifftesting.ReactionFunc{
				rifftesting.InduceFailure("list", "processors"),
			},
			ShouldError: true,
		},
	}

	table.Run(t, commands.NewProcessorListCommand)
}
