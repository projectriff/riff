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

	"github.com/projectriff/riff/cli/pkg/cli"
	"github.com/projectriff/riff/cli/pkg/knative/commands"
	rifftesting "github.com/projectriff/riff/cli/pkg/testing"
	knativev1alpha1 "github.com/projectriff/riff/system/pkg/apis/knative/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestAdapterDeleteOptions(t *testing.T) {
	table := rifftesting.OptionsTable{
		{
			Name: "invalid delete",
			Options: &commands.AdapterDeleteOptions{
				DeleteOptions: rifftesting.InvalidDeleteOptions,
			},
			ExpectFieldErrors: rifftesting.InvalidDeleteOptionsFieldError,
		},
		{
			Name: "valid delete",
			Options: &commands.AdapterDeleteOptions{
				DeleteOptions: rifftesting.ValidDeleteOptions,
			},
			ShouldValidate: true,
		},
	}

	table.Run(t)
}

func TestAdapterDeleteCommand(t *testing.T) {
	adapterName := "test-adapter"
	adapterOtherName := "test-other-adapter"
	defaultNamespace := "default"

	table := rifftesting.CommandTable{
		{
			Name:        "invalid args",
			Args:        []string{},
			ShouldError: true,
		},
		{
			Name: "delete all adapters",
			Args: []string{cli.AllFlagName},
			GivenObjects: []runtime.Object{
				&knativev1alpha1.Adapter{
					ObjectMeta: metav1.ObjectMeta{
						Name:      adapterName,
						Namespace: defaultNamespace,
					},
				},
			},
			ExpectDeleteCollections: []rifftesting.DeleteCollectionRef{{
				Group:     "knative.projectriff.io",
				Resource:  "adapters",
				Namespace: defaultNamespace,
			}},
			ExpectOutput: `
Deleted adapters in namespace "default"
`,
		},
		{
			Name: "delete all adapters error",
			Args: []string{cli.AllFlagName},
			GivenObjects: []runtime.Object{
				&knativev1alpha1.Adapter{
					ObjectMeta: metav1.ObjectMeta{
						Name:      adapterName,
						Namespace: defaultNamespace,
					},
				},
			},
			WithReactors: []rifftesting.ReactionFunc{
				rifftesting.InduceFailure("delete-collection", "adapters"),
			},
			ExpectDeleteCollections: []rifftesting.DeleteCollectionRef{{
				Group:     "knative.projectriff.io",
				Resource:  "adapters",
				Namespace: defaultNamespace,
			}},
			ShouldError: true,
		},
		{
			Name: "delete adapter",
			Args: []string{adapterName},
			GivenObjects: []runtime.Object{
				&knativev1alpha1.Adapter{
					ObjectMeta: metav1.ObjectMeta{
						Name:      adapterName,
						Namespace: defaultNamespace,
					},
				},
			},
			ExpectDeletes: []rifftesting.DeleteRef{{
				Group:     "knative.projectriff.io",
				Resource:  "adapters",
				Namespace: defaultNamespace,
				Name:      adapterName,
			}},
			ExpectOutput: `
Deleted adapter "test-adapter"
`,
		},
		{
			Name: "delete adapters",
			Args: []string{adapterName, adapterOtherName},
			GivenObjects: []runtime.Object{
				&knativev1alpha1.Adapter{
					ObjectMeta: metav1.ObjectMeta{
						Name:      adapterName,
						Namespace: defaultNamespace,
					},
				},
				&knativev1alpha1.Adapter{
					ObjectMeta: metav1.ObjectMeta{
						Name:      adapterOtherName,
						Namespace: defaultNamespace,
					},
				},
			},
			ExpectDeletes: []rifftesting.DeleteRef{{
				Group:     "knative.projectriff.io",
				Resource:  "adapters",
				Namespace: defaultNamespace,
				Name:      adapterName,
			}, {
				Group:     "knative.projectriff.io",
				Resource:  "adapters",
				Namespace: defaultNamespace,
				Name:      adapterOtherName,
			}},
			ExpectOutput: `
Deleted adapter "test-adapter"
Deleted adapter "test-other-adapter"
`,
		},
		{
			Name: "adapter does not exist",
			Args: []string{adapterName},
			ExpectDeletes: []rifftesting.DeleteRef{{
				Group:     "knative.projectriff.io",
				Resource:  "adapters",
				Namespace: defaultNamespace,
				Name:      adapterName,
			}},
			ShouldError: true,
		},
		{
			Name: "delete error",
			Args: []string{adapterName},
			GivenObjects: []runtime.Object{
				&knativev1alpha1.Adapter{
					ObjectMeta: metav1.ObjectMeta{
						Name:      adapterName,
						Namespace: defaultNamespace,
					},
				},
			},
			WithReactors: []rifftesting.ReactionFunc{
				rifftesting.InduceFailure("delete", "adapters"),
			},
			ExpectDeletes: []rifftesting.DeleteRef{{
				Group:     "knative.projectriff.io",
				Resource:  "adapters",
				Namespace: defaultNamespace,
				Name:      adapterName,
			}},
			ShouldError: true,
		},
	}

	table.Run(t, commands.NewAdapterDeleteCommand)
}
