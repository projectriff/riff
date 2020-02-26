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

	"github.com/projectriff/cli/pkg/cli"
	"github.com/projectriff/cli/pkg/streaming/commands"
	rifftesting "github.com/projectriff/cli/pkg/testing"
	streamv1alpha1 "github.com/projectriff/system/pkg/apis/streaming/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestProcessorDeleteOptions(t *testing.T) {
	table := rifftesting.OptionsTable{
		{
			Name: "invalid delete",
			Options: &commands.ProcessorDeleteOptions{
				DeleteOptions: rifftesting.InvalidDeleteOptions,
			},
			ExpectFieldErrors: rifftesting.InvalidDeleteOptionsFieldError,
		},
		{
			Name: "valid delete",
			Options: &commands.ProcessorDeleteOptions{
				DeleteOptions: rifftesting.ValidDeleteOptions,
			},
			ShouldValidate: true,
		},
	}

	table.Run(t)
}

func TestProcessorDeleteCommand(t *testing.T) {
	processorName := "test-processor"
	processorOtherName := "test-other-processor"
	defaultNamespace := "default"

	table := rifftesting.CommandTable{
		{
			Name:        "invalid args",
			Args:        []string{},
			ShouldError: true,
		},
		{
			Name: "delete all processors",
			Args: []string{cli.AllFlagName},
			GivenObjects: []runtime.Object{
				&streamv1alpha1.Processor{
					ObjectMeta: metav1.ObjectMeta{
						Name:      processorName,
						Namespace: defaultNamespace,
					},
				},
			},
			ExpectDeleteCollections: []rifftesting.DeleteCollectionRef{{
				Group:     "streaming.projectriff.io",
				Resource:  "processors",
				Namespace: defaultNamespace,
			}},
			ExpectOutput: `
Deleted processors in namespace "default"
`,
		},
		{
			Name: "delete all processors error",
			Args: []string{cli.AllFlagName},
			GivenObjects: []runtime.Object{
				&streamv1alpha1.Processor{
					ObjectMeta: metav1.ObjectMeta{
						Name:      processorName,
						Namespace: defaultNamespace,
					},
				},
			},
			WithReactors: []rifftesting.ReactionFunc{
				rifftesting.InduceFailure("delete-collection", "processors"),
			},
			ExpectDeleteCollections: []rifftesting.DeleteCollectionRef{{
				Group:     "streaming.projectriff.io",
				Resource:  "processors",
				Namespace: defaultNamespace,
			}},
			ShouldError: true,
		},
		{
			Name: "delete processor",
			Args: []string{processorName},
			GivenObjects: []runtime.Object{
				&streamv1alpha1.Processor{
					ObjectMeta: metav1.ObjectMeta{
						Name:      processorName,
						Namespace: defaultNamespace,
					},
				},
			},
			ExpectDeletes: []rifftesting.DeleteRef{{
				Group:     "streaming.projectriff.io",
				Resource:  "processors",
				Namespace: defaultNamespace,
				Name:      processorName,
			}},
			ExpectOutput: `
Deleted processor "test-processor"
`,
		},
		{
			Name: "delete processors",
			Args: []string{processorName, processorOtherName},
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
						Namespace: defaultNamespace,
					},
				},
			},
			ExpectDeletes: []rifftesting.DeleteRef{{
				Group:     "streaming.projectriff.io",
				Resource:  "processors",
				Namespace: defaultNamespace,
				Name:      processorName,
			}, {
				Group:     "streaming.projectriff.io",
				Resource:  "processors",
				Namespace: defaultNamespace,
				Name:      processorOtherName,
			}},
			ExpectOutput: `
Deleted processor "test-processor"
Deleted processor "test-other-processor"
`,
		},
		{
			Name: "processor does not exist",
			Args: []string{processorName},
			ExpectDeletes: []rifftesting.DeleteRef{{
				Group:     "streaming.projectriff.io",
				Resource:  "processors",
				Namespace: defaultNamespace,
				Name:      processorName,
			}},
			ShouldError: true,
		},
		{
			Name: "delete error",
			Args: []string{processorName},
			GivenObjects: []runtime.Object{
				&streamv1alpha1.Processor{
					ObjectMeta: metav1.ObjectMeta{
						Name:      processorName,
						Namespace: defaultNamespace,
					},
				},
			},
			WithReactors: []rifftesting.ReactionFunc{
				rifftesting.InduceFailure("delete", "processors"),
			},
			ExpectDeletes: []rifftesting.DeleteRef{{
				Group:     "streaming.projectriff.io",
				Resource:  "processors",
				Namespace: defaultNamespace,
				Name:      processorName,
			}},
			ShouldError: true,
		},
	}

	table.Run(t, commands.NewProcessorDeleteCommand)
}
