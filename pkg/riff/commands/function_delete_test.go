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
	"github.com/projectriff/riff/pkg/cli"
	"github.com/projectriff/riff/pkg/riff/commands"
	"github.com/projectriff/riff/pkg/testing"
	buildv1alpha1 "github.com/projectriff/system/pkg/apis/build/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestFunctionDeleteOptions(t *testing.T) {
	table := testing.OptionsTable{
		{
			Name: "invalid delete",
			Options: &commands.FunctionDeleteOptions{
				DeleteOptions: testing.InvalidDeleteOptions,
			},
			ExpectFieldError: testing.InvalidDeleteOptionsFieldError,
		},
		{
			Name: "valid delete",
			Options: &commands.FunctionDeleteOptions{
				DeleteOptions: testing.ValidDeleteOptions,
			},
			ShouldValidate: true,
		},
	}

	table.Run(t)
}

func TestFunctionDeleteCommand(t *testing.T) {
	functionName := "test-function"
	functionOtherName := "test-other-function"
	defaultNamespace := "default"

	table := testing.CommandTable{
		{
			Name:        "invalid args",
			Args:        []string{},
			ShouldError: true,
		},
		{
			Name: "delete all functions",
			Args: []string{cli.AllFlagName},
			GivenObjects: []runtime.Object{
				&buildv1alpha1.Function{
					ObjectMeta: metav1.ObjectMeta{
						Name:      functionName,
						Namespace: defaultNamespace,
					},
				},
			},
			ExpectDeleteCollections: []testing.DeleteCollectionRef{{
				Group:     "build.projectriff.io",
				Resource:  "functions",
				Namespace: defaultNamespace,
			}},
			ExpectOutput: `
Deleted functions in namespace "default"
`,
		},
		{
			Name: "delete function",
			Args: []string{functionName},
			GivenObjects: []runtime.Object{
				&buildv1alpha1.Function{
					ObjectMeta: metav1.ObjectMeta{
						Name:      functionName,
						Namespace: defaultNamespace,
					},
				},
			},
			ExpectDeletes: []testing.DeleteRef{{
				Group:     "build.projectriff.io",
				Resource:  "functions",
				Namespace: defaultNamespace,
				Name:      functionName,
			}},
			ExpectOutput: `
Deleted function "test-function"
`,
		},
		{
			Name: "delete functions",
			Args: []string{functionName, functionOtherName},
			GivenObjects: []runtime.Object{
				&buildv1alpha1.Function{
					ObjectMeta: metav1.ObjectMeta{
						Name:      functionName,
						Namespace: defaultNamespace,
					},
				},
				&buildv1alpha1.Function{
					ObjectMeta: metav1.ObjectMeta{
						Name:      functionOtherName,
						Namespace: defaultNamespace,
					},
				},
			},
			ExpectDeletes: []testing.DeleteRef{{
				Group:     "build.projectriff.io",
				Resource:  "functions",
				Namespace: defaultNamespace,
				Name:      functionName,
			}, {
				Group:     "build.projectriff.io",
				Resource:  "functions",
				Namespace: defaultNamespace,
				Name:      functionOtherName,
			}},
			ExpectOutput: `
Deleted function "test-function"
Deleted function "test-other-function"
`,
		},
		{
			Name: "function does not exist",
			Args: []string{functionName},
			ExpectDeletes: []testing.DeleteRef{{
				Group:     "build.projectriff.io",
				Resource:  "functions",
				Namespace: defaultNamespace,
				Name:      functionName,
			}},
			ShouldError: true,
		},
		{
			Name: "delete error",
			Args: []string{functionName},
			GivenObjects: []runtime.Object{
				&buildv1alpha1.Function{
					ObjectMeta: metav1.ObjectMeta{
						Name:      functionName,
						Namespace: defaultNamespace,
					},
				},
			},
			WithReactors: []testing.ReactionFunc{
				testing.InduceFailure("delete", "functions"),
			},
			ExpectDeletes: []testing.DeleteRef{{
				Group:     "build.projectriff.io",
				Resource:  "functions",
				Namespace: defaultNamespace,
				Name:      functionName,
			}},
			ShouldError: true,
		},
	}

	table.Run(t, commands.NewFunctionDeleteCommand)
}
