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
	"fmt"
	"strings"

	"github.com/projectriff/riff/pkg/cli"
	"github.com/projectriff/riff/pkg/riff/commands"
	"github.com/projectriff/riff/pkg/testing"
	"github.com/projectriff/system/pkg/apis/build"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestCredentialListOptions(t *testing.T) {
	table := testing.OptionsTable{
		{
			Name: "valid list",
			Options: &commands.CredentialListOptions{
				ListOptions: testing.ValidListOptions,
			},
			ShouldValidate: true,
		},
		{
			Name: "invalid list",
			Options: &commands.CredentialListOptions{
				ListOptions: testing.InvalidListOptions,
			},
			ExpectFieldError: testing.InvalidListOptionsFieldError,
		},
	}

	table.Run(t)
}

func TestCredentialListCommand(t *testing.T) {
	t.Parallel()

	credentialName := "test-credential"
	credentialAltName := "test-alt-credential"
	defaultNamespace := "default"
	altNamespace := "alt-namespace"
	credentialLabel := build.CredentialLabelKey

	table := testing.CommandTable{
		{
			Name: "invalid args",
			Args: []string{},
			Prepare: func(t *testing.T, c *cli.Config) error {
				// disable default namespace
				c.Client.(*testing.FakeClient).Namespace = ""
				return nil
			},
			ShouldError: true,
		},
		{
			Name: "empty",
			Args: []string{},
			Verify: func(t *testing.T, output string, err error) {
				if expected, actual := output, "No credentials found.\n"; actual != expected {
					t.Errorf("expected output %q, actually %q", expected, actual)
				}
			},
		},
		{
			Name: "lists a secret",
			Args: []string{},
			GivenObjects: []runtime.Object{
				&corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      credentialName,
						Namespace: defaultNamespace,
						Labels:    map[string]string{credentialLabel: ""},
					},
				},
			},
			Verify: func(t *testing.T, output string, err error) {
				if actual, want := output, fmt.Sprintf("%s\n", credentialName); actual != want {
					t.Errorf("expected output %q, actually %q", want, actual)
				}
			},
		},
		{
			Name: "filters by namespace",
			Args: []string{"--namespace", altNamespace},
			GivenObjects: []runtime.Object{
				&corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      credentialName,
						Namespace: defaultNamespace,
						Labels:    map[string]string{credentialLabel: ""},
					},
				},
			},
			Verify: func(t *testing.T, output string, err error) {
				if actual, want := output, "No credentials found.\n"; actual != want {
					t.Errorf("expected output %q, actually %q", want, actual)
				}
			},
		},
		{
			Name: "all namespace",
			Args: []string{"--all-namespaces"},
			GivenObjects: []runtime.Object{
				&corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      credentialName,
						Namespace: defaultNamespace,
						Labels:    map[string]string{credentialLabel: ""},
					},
				},
				&corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      credentialAltName,
						Namespace: altNamespace,
						Labels:    map[string]string{credentialLabel: ""},
					},
				},
			},
			Verify: func(t *testing.T, output string, err error) {
				for _, expected := range []string{
					fmt.Sprintf("%s\n", credentialName),
					fmt.Sprintf("%s\n", credentialAltName),
				} {
					if !strings.Contains(output, expected) {
						t.Errorf("expected command output to contain %q, actually %q", expected, output)
					}
				}
			},
		},
		{
			Name: "ignore non-riff secrets",
			Args: []string{},
			GivenObjects: []runtime.Object{
				&corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "not-a-credential",
						Namespace: defaultNamespace,
					},
				},
			},
			Verify: func(t *testing.T, output string, err error) {
				if expected, actual := output, "No credentials found.\n"; actual != expected {
					t.Errorf("expected output %q, actually %q", expected, actual)
				}
			},
		},
		{
			Name: "list error",
			Args: []string{},
			WithReactors: []testing.ReactionFunc{
				testing.InduceFailure("list", "secrets"),
			},
			ShouldError: true,
		},
	}

	table.Run(t, commands.NewCredentialListCommand)
}
