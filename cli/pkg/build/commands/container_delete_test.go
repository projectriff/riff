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

	"github.com/projectriff/cli/pkg/build/commands"
	"github.com/projectriff/cli/pkg/cli"
	rifftesting "github.com/projectriff/cli/pkg/testing"
	buildv1alpha1 "github.com/projectriff/system/pkg/apis/build/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestContainerDeleteOptions(t *testing.T) {
	table := rifftesting.OptionsTable{
		{
			Name: "invalid delete",
			Options: &commands.ContainerDeleteOptions{
				DeleteOptions: rifftesting.InvalidDeleteOptions,
			},
			ExpectFieldErrors: rifftesting.InvalidDeleteOptionsFieldError,
		},
		{
			Name: "valid delete",
			Options: &commands.ContainerDeleteOptions{
				DeleteOptions: rifftesting.ValidDeleteOptions,
			},
			ShouldValidate: true,
		},
	}

	table.Run(t)
}

func TestContainerDeleteCommand(t *testing.T) {
	containerName := "test-container"
	containerOtherName := "test-other-container"
	defaultNamespace := "default"

	table := rifftesting.CommandTable{
		{
			Name:        "invalid args",
			Args:        []string{},
			ShouldError: true,
		},
		{
			Name: "delete all containers",
			Args: []string{cli.AllFlagName},
			GivenObjects: []runtime.Object{
				&buildv1alpha1.Container{
					ObjectMeta: metav1.ObjectMeta{
						Name:      containerName,
						Namespace: defaultNamespace,
					},
				},
			},
			ExpectDeleteCollections: []rifftesting.DeleteCollectionRef{{
				Group:     "build.projectriff.io",
				Resource:  "containers",
				Namespace: defaultNamespace,
			}},
			ExpectOutput: `
Deleted containers in namespace "default"
`,
		},
		{
			Name: "delete all containers error",
			Args: []string{cli.AllFlagName},
			GivenObjects: []runtime.Object{
				&buildv1alpha1.Container{
					ObjectMeta: metav1.ObjectMeta{
						Name:      containerName,
						Namespace: defaultNamespace,
					},
				},
			},
			WithReactors: []rifftesting.ReactionFunc{
				rifftesting.InduceFailure("delete-collection", "containers"),
			},
			ExpectDeleteCollections: []rifftesting.DeleteCollectionRef{{
				Group:     "build.projectriff.io",
				Resource:  "containers",
				Namespace: defaultNamespace,
			}},
			ShouldError: true,
		},
		{
			Name: "delete container",
			Args: []string{containerName},
			GivenObjects: []runtime.Object{
				&buildv1alpha1.Container{
					ObjectMeta: metav1.ObjectMeta{
						Name:      containerName,
						Namespace: defaultNamespace,
					},
				},
			},
			ExpectDeletes: []rifftesting.DeleteRef{{
				Group:     "build.projectriff.io",
				Resource:  "containers",
				Namespace: defaultNamespace,
				Name:      containerName,
			}},
			ExpectOutput: `
Deleted container "test-container"
`,
		},
		{
			Name: "delete containers",
			Args: []string{containerName, containerOtherName},
			GivenObjects: []runtime.Object{
				&buildv1alpha1.Container{
					ObjectMeta: metav1.ObjectMeta{
						Name:      containerName,
						Namespace: defaultNamespace,
					},
				},
				&buildv1alpha1.Container{
					ObjectMeta: metav1.ObjectMeta{
						Name:      containerOtherName,
						Namespace: defaultNamespace,
					},
				},
			},
			ExpectDeletes: []rifftesting.DeleteRef{{
				Group:     "build.projectriff.io",
				Resource:  "containers",
				Namespace: defaultNamespace,
				Name:      containerName,
			}, {
				Group:     "build.projectriff.io",
				Resource:  "containers",
				Namespace: defaultNamespace,
				Name:      containerOtherName,
			}},
			ExpectOutput: `
Deleted container "test-container"
Deleted container "test-other-container"
`,
		},
		{
			Name: "container does not exist",
			Args: []string{containerName},
			ExpectDeletes: []rifftesting.DeleteRef{{
				Group:     "build.projectriff.io",
				Resource:  "containers",
				Namespace: defaultNamespace,
				Name:      containerName,
			}},
			ShouldError: true,
		},
		{
			Name: "delete error",
			Args: []string{containerName},
			GivenObjects: []runtime.Object{
				&buildv1alpha1.Container{
					ObjectMeta: metav1.ObjectMeta{
						Name:      containerName,
						Namespace: defaultNamespace,
					},
				},
			},
			WithReactors: []rifftesting.ReactionFunc{
				rifftesting.InduceFailure("delete", "containers"),
			},
			ExpectDeletes: []rifftesting.DeleteRef{{
				Group:     "build.projectriff.io",
				Resource:  "containers",
				Namespace: defaultNamespace,
				Name:      containerName,
			}},
			ShouldError: true,
		},
	}

	table.Run(t, commands.NewContainerDeleteCommand)
}
