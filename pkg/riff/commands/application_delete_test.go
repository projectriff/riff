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

	"github.com/projectriff/riff/pkg/cli"
	"github.com/projectriff/riff/pkg/riff/commands"
	rifftesting "github.com/projectriff/riff/pkg/testing"
	buildv1alpha1 "github.com/projectriff/system/pkg/apis/build/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestApplicationDeleteOptions(t *testing.T) {
	table := rifftesting.OptionsTable{
		{
			Name: "invalid delete",
			Options: &commands.ApplicationDeleteOptions{
				DeleteOptions: rifftesting.InvalidDeleteOptions,
			},
			ExpectFieldError: rifftesting.InvalidDeleteOptionsFieldError,
		},
		{
			Name: "valid delete",
			Options: &commands.ApplicationDeleteOptions{
				DeleteOptions: rifftesting.ValidDeleteOptions,
			},
			ShouldValidate: true,
		},
	}

	table.Run(t)
}

func TestApplicationDeleteCommand(t *testing.T) {
	applicationName := "test-application"
	applicationOtherName := "test-other-application"
	defaultNamespace := "default"

	table := rifftesting.CommandTable{
		{
			Name:        "invalid args",
			Args:        []string{},
			ShouldError: true,
		},
		{
			Name: "delete all applications",
			Args: []string{cli.AllFlagName},
			GivenObjects: []runtime.Object{
				&buildv1alpha1.Application{
					ObjectMeta: metav1.ObjectMeta{
						Name:      applicationName,
						Namespace: defaultNamespace,
					},
				},
			},
			ExpectDeleteCollections: []rifftesting.DeleteCollectionRef{{
				Group:     "build.projectriff.io",
				Resource:  "applications",
				Namespace: defaultNamespace,
			}},
			ExpectOutput: `
Deleted applications in namespace "default"
`,
		},
		{
			Name: "delete all applications error",
			Args: []string{cli.AllFlagName},
			GivenObjects: []runtime.Object{
				&buildv1alpha1.Application{
					ObjectMeta: metav1.ObjectMeta{
						Name:      applicationName,
						Namespace: defaultNamespace,
					},
				},
			},
			WithReactors: []rifftesting.ReactionFunc{
				rifftesting.InduceFailure("delete-collection", "applications"),
			},
			ExpectDeleteCollections: []rifftesting.DeleteCollectionRef{{
				Group:     "build.projectriff.io",
				Resource:  "applications",
				Namespace: defaultNamespace,
			}},
			ShouldError: true,
		},
		{
			Name: "delete application",
			Args: []string{applicationName},
			GivenObjects: []runtime.Object{
				&buildv1alpha1.Application{
					ObjectMeta: metav1.ObjectMeta{
						Name:      applicationName,
						Namespace: defaultNamespace,
					},
				},
			},
			ExpectDeletes: []rifftesting.DeleteRef{{
				Group:     "build.projectriff.io",
				Resource:  "applications",
				Namespace: defaultNamespace,
				Name:      applicationName,
			}},
			ExpectOutput: `
Deleted application "test-application"
`,
		},
		{
			Name: "delete applications",
			Args: []string{applicationName, applicationOtherName},
			GivenObjects: []runtime.Object{
				&buildv1alpha1.Application{
					ObjectMeta: metav1.ObjectMeta{
						Name:      applicationName,
						Namespace: defaultNamespace,
					},
				},
				&buildv1alpha1.Application{
					ObjectMeta: metav1.ObjectMeta{
						Name:      applicationOtherName,
						Namespace: defaultNamespace,
					},
				},
			},
			ExpectDeletes: []rifftesting.DeleteRef{{
				Group:     "build.projectriff.io",
				Resource:  "applications",
				Namespace: defaultNamespace,
				Name:      applicationName,
			}, {
				Group:     "build.projectriff.io",
				Resource:  "applications",
				Namespace: defaultNamespace,
				Name:      applicationOtherName,
			}},
			ExpectOutput: `
Deleted application "test-application"
Deleted application "test-other-application"
`,
		},
		{
			Name: "application does not exist",
			Args: []string{applicationName},
			ExpectDeletes: []rifftesting.DeleteRef{{
				Group:     "build.projectriff.io",
				Resource:  "applications",
				Namespace: defaultNamespace,
				Name:      applicationName,
			}},
			ShouldError: true,
		},
		{
			Name: "delete error",
			Args: []string{applicationName},
			GivenObjects: []runtime.Object{
				&buildv1alpha1.Application{
					ObjectMeta: metav1.ObjectMeta{
						Name:      applicationName,
						Namespace: defaultNamespace,
					},
				},
			},
			WithReactors: []rifftesting.ReactionFunc{
				rifftesting.InduceFailure("delete", "applications"),
			},
			ExpectDeletes: []rifftesting.DeleteRef{{
				Group:     "build.projectriff.io",
				Resource:  "applications",
				Namespace: defaultNamespace,
				Name:      applicationName,
			}},
			ShouldError: true,
		},
	}

	table.Run(t, commands.NewApplicationDeleteCommand)
}
