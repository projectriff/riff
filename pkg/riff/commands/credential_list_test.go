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
	credentialName := "test-credential"
	credentialOtherName := "test-other-credential"
	defaultNamespace := "default"
	otherNamespace := "other-namespace"
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
			Name:         "empty",
			Args:         []string{},
			ExpectOutput: "No credentials found.\n",
		},
		{
			Name: "lists an item",
			Args: []string{},
			GivenObjects: []runtime.Object{
				&corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:        credentialName,
						Namespace:   defaultNamespace,
						Labels:      map[string]string{credentialLabel: "docker-hub"},
						Annotations: map[string]string{"build.knative.dev/docker-0": "https://index.docker.io/v1/"},
					},
				},
			},
			ExpectOutput: `
NAME              TYPE         REGISTRY                      AGE
test-credential   docker-hub   https://index.docker.io/v1/   <unknown>
`,
		},
		{
			Name: "filters by namespace",
			Args: []string{cli.NamespaceFlagName, otherNamespace},
			GivenObjects: []runtime.Object{
				&corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:        credentialName,
						Namespace:   defaultNamespace,
						Labels:      map[string]string{credentialLabel: "docker-hub"},
						Annotations: map[string]string{"build.knative.dev/docker-0": "https://index.docker.io/v1/"},
					},
				},
			},
			ExpectOutput: "No credentials found.\n",
		},
		{
			Name: "table populates all columns",
			Args: []string{},
			GivenObjects: []runtime.Object{
				&corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:        "registry",
						Namespace:   defaultNamespace,
						Labels:      map[string]string{credentialLabel: "basic-auth"},
						Annotations: map[string]string{"build.knative.dev/docker-0": "https://registry.example.com/"},
					},
				},
				&corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:        "docker-hub",
						Namespace:   defaultNamespace,
						Labels:      map[string]string{credentialLabel: "docker-hub"},
						Annotations: map[string]string{"build.knative.dev/docker-0": "https://index.docker.io/v1/"},
					},
				},
				&corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:        "gcr",
						Namespace:   defaultNamespace,
						Labels:      map[string]string{credentialLabel: "gcr"},
						Annotations: map[string]string{"build.knative.dev/docker-0": "https://gcr.io"},
					},
				},
			},
			ExpectOutput: `
NAME         TYPE         REGISTRY                        AGE
docker-hub   docker-hub   https://index.docker.io/v1/     <unknown>
gcr          gcr          https://gcr.io                  <unknown>
registry     basic-auth   https://registry.example.com/   <unknown>
`,
		},
		{
			Name: "all namespace",
			Args: []string{cli.AllNamespacesFlagName},
			GivenObjects: []runtime.Object{
				&corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:        credentialName,
						Namespace:   defaultNamespace,
						Labels:      map[string]string{credentialLabel: "docker-hub"},
						Annotations: map[string]string{"build.knative.dev/docker-0": "https://index.docker.io/v1/"},
					},
				},
				&corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:        credentialOtherName,
						Namespace:   otherNamespace,
						Labels:      map[string]string{credentialLabel: "docker-hub"},
						Annotations: map[string]string{"build.knative.dev/docker-0": "https://index.docker.io/v1/"},
					},
				},
			},
			ExpectOutput: `
NAMESPACE         NAME                    TYPE         REGISTRY                      AGE
default           test-credential         docker-hub   https://index.docker.io/v1/   <unknown>
other-namespace   test-other-credential   docker-hub   https://index.docker.io/v1/   <unknown>
`,
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
			ExpectOutput: "No credentials found.\n",
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
