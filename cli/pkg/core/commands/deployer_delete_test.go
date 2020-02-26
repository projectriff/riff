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
	"github.com/projectriff/cli/pkg/core/commands"
	rifftesting "github.com/projectriff/cli/pkg/testing"
	corev1alpha1 "github.com/projectriff/system/pkg/apis/core/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestDeployerDeleteOptions(t *testing.T) {
	table := rifftesting.OptionsTable{
		{
			Name: "invalid delete",
			Options: &commands.DeployerDeleteOptions{
				DeleteOptions: rifftesting.InvalidDeleteOptions,
			},
			ExpectFieldErrors: rifftesting.InvalidDeleteOptionsFieldError,
		},
		{
			Name: "valid delete",
			Options: &commands.DeployerDeleteOptions{
				DeleteOptions: rifftesting.ValidDeleteOptions,
			},
			ShouldValidate: true,
		},
	}

	table.Run(t)
}

func TestDeployerDeleteCommand(t *testing.T) {
	deployerName := "test-deployer"
	deployerOtherName := "test-other-deployer"
	defaultNamespace := "default"

	table := rifftesting.CommandTable{
		{
			Name:        "invalid args",
			Args:        []string{},
			ShouldError: true,
		},
		{
			Name: "delete all deployers",
			Args: []string{cli.AllFlagName},
			GivenObjects: []runtime.Object{
				&corev1alpha1.Deployer{
					ObjectMeta: metav1.ObjectMeta{
						Name:      deployerName,
						Namespace: defaultNamespace,
					},
				},
			},
			ExpectDeleteCollections: []rifftesting.DeleteCollectionRef{{
				Group:     "core.projectriff.io",
				Resource:  "deployers",
				Namespace: defaultNamespace,
			}},
			ExpectOutput: `
Deleted deployers in namespace "default"
`,
		},
		{
			Name: "delete all deployers error",
			Args: []string{cli.AllFlagName},
			GivenObjects: []runtime.Object{
				&corev1alpha1.Deployer{
					ObjectMeta: metav1.ObjectMeta{
						Name:      deployerName,
						Namespace: defaultNamespace,
					},
				},
			},
			WithReactors: []rifftesting.ReactionFunc{
				rifftesting.InduceFailure("delete-collection", "deployers"),
			},
			ExpectDeleteCollections: []rifftesting.DeleteCollectionRef{{
				Group:     "core.projectriff.io",
				Resource:  "deployers",
				Namespace: defaultNamespace,
			}},
			ShouldError: true,
		},
		{
			Name: "delete deployer",
			Args: []string{deployerName},
			GivenObjects: []runtime.Object{
				&corev1alpha1.Deployer{
					ObjectMeta: metav1.ObjectMeta{
						Name:      deployerName,
						Namespace: defaultNamespace,
					},
				},
			},
			ExpectDeletes: []rifftesting.DeleteRef{{
				Group:     "core.projectriff.io",
				Resource:  "deployers",
				Namespace: defaultNamespace,
				Name:      deployerName,
			}},
			ExpectOutput: `
Deleted deployer "test-deployer"
`,
		},
		{
			Name: "delete deployers",
			Args: []string{deployerName, deployerOtherName},
			GivenObjects: []runtime.Object{
				&corev1alpha1.Deployer{
					ObjectMeta: metav1.ObjectMeta{
						Name:      deployerName,
						Namespace: defaultNamespace,
					},
				},
				&corev1alpha1.Deployer{
					ObjectMeta: metav1.ObjectMeta{
						Name:      deployerOtherName,
						Namespace: defaultNamespace,
					},
				},
			},
			ExpectDeletes: []rifftesting.DeleteRef{{
				Group:     "core.projectriff.io",
				Resource:  "deployers",
				Namespace: defaultNamespace,
				Name:      deployerName,
			}, {
				Group:     "core.projectriff.io",
				Resource:  "deployers",
				Namespace: defaultNamespace,
				Name:      deployerOtherName,
			}},
			ExpectOutput: `
Deleted deployer "test-deployer"
Deleted deployer "test-other-deployer"
`,
		},
		{
			Name: "deployer does not exist",
			Args: []string{deployerName},
			ExpectDeletes: []rifftesting.DeleteRef{{
				Group:     "core.projectriff.io",
				Resource:  "deployers",
				Namespace: defaultNamespace,
				Name:      deployerName,
			}},
			ShouldError: true,
		},
		{
			Name: "delete error",
			Args: []string{deployerName},
			GivenObjects: []runtime.Object{
				&corev1alpha1.Deployer{
					ObjectMeta: metav1.ObjectMeta{
						Name:      deployerName,
						Namespace: defaultNamespace,
					},
				},
			},
			WithReactors: []rifftesting.ReactionFunc{
				rifftesting.InduceFailure("delete", "deployers"),
			},
			ExpectDeletes: []rifftesting.DeleteRef{{
				Group:     "core.projectriff.io",
				Resource:  "deployers",
				Namespace: defaultNamespace,
				Name:      deployerName,
			}},
			ShouldError: true,
		},
	}

	table.Run(t, commands.NewDeployerDeleteCommand)
}
