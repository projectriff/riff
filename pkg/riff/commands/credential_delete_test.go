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
	"github.com/projectriff/system/pkg/apis/build"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestCredentialDeleteOptions(t *testing.T) {
	table := testing.OptionsTable{
		{
			Name: "valid multi-delete",
			Options: &commands.CredentialDeleteOptions{
				DeleteOptions: testing.ValidDeleteOptions,
			},
			ShouldValidate: true,
		},
		{
			Name: "invalid multi-delete",
			Options: &commands.CredentialDeleteOptions{
				DeleteOptions: testing.InvalidDeleteOptions,
			},
			ExpectFieldError: testing.InvalidDeleteOptionsFieldError,
		},
	}

	table.Run(t)
}

func TestCredentialDeleteCommand(t *testing.T) {
	credentialName := "test-credential"
	credentialOtherName := "test-other-credential"
	defaultNamespace := "default"
	credentialLabel := build.CredentialLabelKey

	table := testing.CommandTable{
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
			ExpectDeleteCollections: []testing.DeleteCollectionRef{{
				Resource:      "secrets",
				Namespace:     defaultNamespace,
				LabelSelector: credentialLabel,
			}},
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
			WithReactors: []testing.ReactionFunc{
				testing.InduceFailure("delete-collection", "secrets"),
			},
			ExpectDeleteCollections: []testing.DeleteCollectionRef{{
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
			ExpectDeletes: []testing.DeleteRef{{
				Resource:  "secrets",
				Namespace: defaultNamespace,
				Name:      credentialName,
			}},
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
			ExpectDeletes: []testing.DeleteRef{{
				Resource:  "secrets",
				Namespace: defaultNamespace,
				Name:      credentialName,
			}, {
				Resource:  "secrets",
				Namespace: defaultNamespace,
				Name:      credentialOtherName,
			}},
		},
		{
			Name: "secret does not exist",
			Args: []string{credentialName},
			ExpectDeletes: []testing.DeleteRef{{
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
			WithReactors: []testing.ReactionFunc{
				testing.InduceFailure("delete", "secrets"),
			},
			ExpectDeletes: []testing.DeleteRef{{
				Resource:  "secrets",
				Namespace: defaultNamespace,
				Name:      credentialName,
			}},
			ShouldError: true,
		},
	}

	table.Run(t, commands.NewCredentialDeleteCommand)
}
