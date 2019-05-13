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
	requestv1alpha1 "github.com/projectriff/system/pkg/apis/request/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestRequestProcessorDeleteOptions(t *testing.T) {
	table := testing.OptionsTable{
		{
			Name: "invalid delete",
			Options: &commands.RequestProcessorDeleteOptions{
				DeleteOptions: testing.InvalidDeleteOptions,
			},
			ExpectFieldError: testing.InvalidDeleteOptionsFieldError,
		},
		{
			Name: "valid delete",
			Options: &commands.RequestProcessorDeleteOptions{
				DeleteOptions: testing.ValidDeleteOptions,
			},
			ShouldValidate: true,
		},
	}

	table.Run(t)
}

func TestRequestProcessorDeleteCommand(t *testing.T) {
	t.Parallel()

	requestprocessorName := "test-function"
	requestprocessorAltName := "test-alt-function"
	defaultNamespace := "default"

	table := testing.CommandTable{
		{
			Name:        "invalid args",
			Args:        []string{},
			ShouldError: true,
		},
		{
			Name: "delete all requestprocessors",
			Args: []string{"--all"},
			GivenObjects: []runtime.Object{
				&requestv1alpha1.RequestProcessor{
					ObjectMeta: metav1.ObjectMeta{
						Name:      requestprocessorName,
						Namespace: defaultNamespace,
					},
				},
			},
			ExpectDeleteCollections: []testing.DeleteCollectionRef{{
				Group:     "request.projectriff.io",
				Resource:  "requestprocessors",
				Namespace: defaultNamespace,
			}},
		},
		{
			Name: "delete request processor",
			Args: []string{requestprocessorName},
			GivenObjects: []runtime.Object{
				&requestv1alpha1.RequestProcessor{
					ObjectMeta: metav1.ObjectMeta{
						Name:      requestprocessorName,
						Namespace: defaultNamespace,
					},
				},
			},
			ExpectDeletes: []testing.DeleteRef{{
				Group:     "request.projectriff.io",
				Resource:  "requestprocessors",
				Namespace: defaultNamespace,
				Name:      requestprocessorName,
			}},
		},
		{
			Name: "delete request processors",
			Args: []string{requestprocessorName, requestprocessorAltName},
			GivenObjects: []runtime.Object{
				&requestv1alpha1.RequestProcessor{
					ObjectMeta: metav1.ObjectMeta{
						Name:      requestprocessorName,
						Namespace: defaultNamespace,
					},
				},
				&requestv1alpha1.RequestProcessor{
					ObjectMeta: metav1.ObjectMeta{
						Name:      requestprocessorAltName,
						Namespace: defaultNamespace,
					},
				},
			},
			ExpectDeletes: []testing.DeleteRef{{
				Group:     "request.projectriff.io",
				Resource:  "requestprocessors",
				Namespace: defaultNamespace,
				Name:      requestprocessorName,
			}, {
				Group:     "request.projectriff.io",
				Resource:  "requestprocessors",
				Namespace: defaultNamespace,
				Name:      requestprocessorAltName,
			}},
		},
		{
			Name: "request processor does not exist",
			Args: []string{requestprocessorName},
			ExpectDeletes: []testing.DeleteRef{{
				Group:     "request.projectriff.io",
				Resource:  "requestprocessors",
				Namespace: defaultNamespace,
				Name:      requestprocessorName,
			}},
			ShouldError: true,
		},
		{
			Name: "delete error",
			Args: []string{requestprocessorName},
			GivenObjects: []runtime.Object{
				&requestv1alpha1.RequestProcessor{
					ObjectMeta: metav1.ObjectMeta{
						Name:      requestprocessorName,
						Namespace: defaultNamespace,
					},
				},
			},
			WithReactors: []testing.ReactionFunc{
				testing.InduceFailure("delete", "requestprocessors"),
			},
			ExpectDeletes: []testing.DeleteRef{{
				Group:     "request.projectriff.io",
				Resource:  "requestprocessors",
				Namespace: defaultNamespace,
				Name:      requestprocessorName,
			}},
			ShouldError: true,
		},
	}

	table.Run(t, commands.NewRequestProcessorDeleteCommand)
}
