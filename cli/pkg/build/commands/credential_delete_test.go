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

	"github.com/projectriff/riff/cli/pkg/build/commands"
	"github.com/projectriff/riff/cli/pkg/cli"
	rifftesting "github.com/projectriff/riff/cli/pkg/testing"
	buildv1alpha1 "github.com/projectriff/system/pkg/apis/build/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestCredentialDeleteOptions(t *testing.T) {
	table := rifftesting.OptionsTable{
		{
			Name: "valid multi-delete",
			Options: &commands.CredentialDeleteOptions{
				DeleteOptions: rifftesting.ValidDeleteOptions,
			},
			ShouldValidate: true,
		},
		{
			Name: "invalid multi-delete",
			Options: &commands.CredentialDeleteOptions{
				DeleteOptions: rifftesting.InvalidDeleteOptions,
			},
			ExpectFieldErrors: rifftesting.InvalidDeleteOptionsFieldError,
		},
	}

	table.Run(t)
}

func TestCredentialDeleteCommand(t *testing.T) {
	credentialName := "test-credential"
	credentialOtherName := "test-other-credential"
	defaultNamespace := "default"
	credentialLabel := buildv1alpha1.CredentialLabelKey

	table := rifftesting.CommandTable{
		{
			Name:        "invalid args",
			Args:        []string{},
			ShouldError: true,
		},
		{
			Name: "delete all secrets",
			Args: []string{cli.AllFlagName},
			GivenObjects: []runtime.Object{
				&corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      credentialName,
						Namespace: defaultNamespace,
						Labels:    map[string]string{credentialLabel: ""},
					},
					StringData: map[string]string{},
				},
			},
			ExpectDeleteCollections: []rifftesting.DeleteCollectionRef{{
				Resource:      "secrets",
				Namespace:     defaultNamespace,
				LabelSelector: credentialLabel,
			}},
			ExpectOutput: `
Deleted credentials in namespace "default"
`,
		},
		{
			Name: "delete all secrets error",
			Args: []string{cli.AllFlagName},
			GivenObjects: []runtime.Object{
				&corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      credentialName,
						Namespace: defaultNamespace,
						Labels:    map[string]string{credentialLabel: ""},
					},
					StringData: map[string]string{},
				},
			},
			WithReactors: []rifftesting.ReactionFunc{
				rifftesting.InduceFailure("delete-collection", "secrets"),
			},
			ExpectDeleteCollections: []rifftesting.DeleteCollectionRef{{
				Resource:      "secrets",
				Namespace:     defaultNamespace,
				LabelSelector: credentialLabel,
			}},
			ShouldError: true,
		},
		{
			Name: "delete secret",
			Args: []string{credentialName},
			GivenObjects: []runtime.Object{
				&corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      credentialName,
						Namespace: defaultNamespace,
						Labels:    map[string]string{credentialLabel: ""},
					},
					StringData: map[string]string{},
				},
			},
			ExpectDeletes: []rifftesting.DeleteRef{{
				Resource:  "secrets",
				Namespace: defaultNamespace,
				Name:      credentialName,
			}},
			ExpectOutput: `
Deleted credential "test-credential"
`,
		},
		{
			Name: "delete secrets",
			Args: []string{credentialName, credentialOtherName},
			GivenObjects: []runtime.Object{
				&corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      credentialName,
						Namespace: defaultNamespace,
						Labels:    map[string]string{credentialLabel: ""},
					},
					StringData: map[string]string{},
				},
				&corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      credentialOtherName,
						Namespace: defaultNamespace,
						Labels:    map[string]string{credentialLabel: ""},
					},
					StringData: map[string]string{},
				},
			},
			ExpectDeletes: []rifftesting.DeleteRef{{
				Resource:  "secrets",
				Namespace: defaultNamespace,
				Name:      credentialName,
			}, {
				Resource:  "secrets",
				Namespace: defaultNamespace,
				Name:      credentialOtherName,
			}},
			ExpectOutput: `
Deleted credential "test-credential"
Deleted credential "test-other-credential"
`,
		},
		{
			Name: "secret does not exist",
			Args: []string{credentialName},
			ExpectDeletes: []rifftesting.DeleteRef{{
				Resource:  "secrets",
				Namespace: defaultNamespace,
				Name:      credentialName,
			}},
			ShouldError: true,
		},
		{
			Name: "delete error",
			Args: []string{credentialName},
			GivenObjects: []runtime.Object{
				&corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      credentialName,
						Namespace: defaultNamespace,
						Labels:    map[string]string{credentialLabel: ""},
					},
					StringData: map[string]string{},
				},
			},
			WithReactors: []rifftesting.ReactionFunc{
				rifftesting.InduceFailure("delete", "secrets"),
			},
			ExpectDeletes: []rifftesting.DeleteRef{{
				Resource:  "secrets",
				Namespace: defaultNamespace,
				Name:      credentialName,
			}},
			ShouldError: true,
		},
	}

	table.Run(t, commands.NewCredentialDeleteCommand)
}
