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
	requestv1alpha1 "github.com/projectriff/system/pkg/apis/request/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestHandlerDeleteOptions(t *testing.T) {
	table := testing.OptionsTable{
		{
			Name: "invalid delete",
			Options: &commands.HandlerDeleteOptions{
				DeleteOptions: testing.InvalidDeleteOptions,
			},
			ExpectFieldError: testing.InvalidDeleteOptionsFieldError,
		},
		{
			Name: "valid delete",
			Options: &commands.HandlerDeleteOptions{
				DeleteOptions: testing.ValidDeleteOptions,
			},
			ShouldValidate: true,
		},
	}

	table.Run(t)
}

func TestHandlerDeleteCommand(t *testing.T) {
	handlerName := "test-handler"
	handlerOtherName := "test-other-handler"
	defaultNamespace := "default"

	table := testing.CommandTable{
		{
			Name:        "invalid args",
			Args:        []string{},
			ShouldError: true,
		},
		{
			Name: "delete all handlers",
			Args: []string{cli.AllFlagName},
			GivenObjects: []runtime.Object{
				&requestv1alpha1.Handler{
					ObjectMeta: metav1.ObjectMeta{
						Name:      handlerName,
						Namespace: defaultNamespace,
					},
				},
			},
			ExpectDeleteCollections: []testing.DeleteCollectionRef{{
				Group:     "request.projectriff.io",
				Resource:  "handlers",
				Namespace: defaultNamespace,
			}},
			ExpectOutput: `
Deleted handlers in namespace "default"
`,
		},
		{
			Name: "delete handler",
			Args: []string{handlerName},
			GivenObjects: []runtime.Object{
				&requestv1alpha1.Handler{
					ObjectMeta: metav1.ObjectMeta{
						Name:      handlerName,
						Namespace: defaultNamespace,
					},
				},
			},
			ExpectDeletes: []testing.DeleteRef{{
				Group:     "request.projectriff.io",
				Resource:  "handlers",
				Namespace: defaultNamespace,
				Name:      handlerName,
			}},
			ExpectOutput: `
Deleted handler "test-handler"
`,
		},
		{
			Name: "delete handlers",
			Args: []string{handlerName, handlerOtherName},
			GivenObjects: []runtime.Object{
				&requestv1alpha1.Handler{
					ObjectMeta: metav1.ObjectMeta{
						Name:      handlerName,
						Namespace: defaultNamespace,
					},
				},
				&requestv1alpha1.Handler{
					ObjectMeta: metav1.ObjectMeta{
						Name:      handlerOtherName,
						Namespace: defaultNamespace,
					},
				},
			},
			ExpectDeletes: []testing.DeleteRef{{
				Group:     "request.projectriff.io",
				Resource:  "handlers",
				Namespace: defaultNamespace,
				Name:      handlerName,
			}, {
				Group:     "request.projectriff.io",
				Resource:  "handlers",
				Namespace: defaultNamespace,
				Name:      handlerOtherName,
			}},
			ExpectOutput: `
Deleted handler "test-handler"
Deleted handler "test-other-handler"
`,
		},
		{
			Name: "handler does not exist",
			Args: []string{handlerName},
			ExpectDeletes: []testing.DeleteRef{{
				Group:     "request.projectriff.io",
				Resource:  "handlers",
				Namespace: defaultNamespace,
				Name:      handlerName,
			}},
			ShouldError: true,
		},
		{
			Name: "delete error",
			Args: []string{handlerName},
			GivenObjects: []runtime.Object{
				&requestv1alpha1.Handler{
					ObjectMeta: metav1.ObjectMeta{
						Name:      handlerName,
						Namespace: defaultNamespace,
					},
				},
			},
			WithReactors: []testing.ReactionFunc{
				testing.InduceFailure("delete", "handlers"),
			},
			ExpectDeletes: []testing.DeleteRef{{
				Group:     "request.projectriff.io",
				Resource:  "handlers",
				Namespace: defaultNamespace,
				Name:      handlerName,
			}},
			ShouldError: true,
		},
	}

	table.Run(t, commands.NewHandlerDeleteCommand)
}
