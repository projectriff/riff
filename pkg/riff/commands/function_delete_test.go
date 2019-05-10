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
	"github.com/projectriff/riff/pkg/riff/commands"
	"github.com/projectriff/riff/pkg/testing"
	buildv1alpha1 "github.com/projectriff/system/pkg/apis/build/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestFunctionDeleteCommand(t *testing.T) {
	t.Parallel()

	functionName := "test-function"
	functionAltName := "test-alt-function"
	defaultNamespace := "default"

	table := testing.CommandTable{
		{
			Name: "delete all functions",
			Args: []string{"--all"},
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
			ExpectDeleteCollections: []testing.DeleteCollectionRef{{
				Group:     "build.projectriff.io",
				Resource:  "functions",
				Namespace: defaultNamespace,
			}},
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
		},
		{
			Name: "delete functions",
			Args: []string{functionName, functionAltName},
			GivenObjects: []runtime.Object{
				&buildv1alpha1.Function{
					ObjectMeta: metav1.ObjectMeta{
						Name:      functionName,
						Namespace: defaultNamespace,
					},
				},
				&buildv1alpha1.Function{
					ObjectMeta: metav1.ObjectMeta{
						Name:      functionAltName,
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
				Name:      functionAltName,
			}},
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
