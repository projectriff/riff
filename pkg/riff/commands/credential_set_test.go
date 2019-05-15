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

func TestCredentialSetOptions(t *testing.T) {
	table := testing.OptionsTable{
		{
			Name: "valid namespaced resource",
			Options: &commands.CredentialSetOptions{
				ResourceOptions: testing.ValidResourceOptions,
			},
			ExpectFieldError: cli.ErrMissingOneOf("docker-hub", "gcr", "registry"),
		},
		{
			Name: "invalid namespaced resource",
			Options: &commands.CredentialSetOptions{
				ResourceOptions: testing.InvalidResourceOptions,
			},
			ExpectFieldError: testing.InvalidResourceOptionsFieldError.Also(
				cli.ErrMissingOneOf("docker-hub", "gcr", "registry"),
			),
		},
		{
			Name: "docker hub",
			Options: &commands.CredentialSetOptions{
				ResourceOptions:   testing.ValidResourceOptions,
				DockerHubId:       "projectriff",
				DockerHubPassword: "1password",
			},
			ShouldValidate: true,
		},
		{
			Name: "docker hub missing password",
			Options: &commands.CredentialSetOptions{
				ResourceOptions: testing.ValidResourceOptions,
				DockerHubId:     "projectriff",
			},
			ExpectFieldError: cli.ErrMissingField("docker-hub-password"),
		},
		{
			Name: "gcr",
			Options: &commands.CredentialSetOptions{
				ResourceOptions: testing.ValidResourceOptions,
				GcrTokenPath:    "gcr-credentials.json",
			},
			ShouldValidate: true,
		},
		{
			Name: "registry",
			Options: &commands.CredentialSetOptions{
				ResourceOptions:  testing.ValidResourceOptions,
				Registry:         "example.com",
				RegistryUser:     "projectriff",
				RegistryPassword: "1password",
			},
			ShouldValidate: true,
		},
		{
			Name: "registry missing user",
			Options: &commands.CredentialSetOptions{
				ResourceOptions:  testing.ValidResourceOptions,
				Registry:         "example.com",
				RegistryPassword: "1password",
			},
			ExpectFieldError: cli.ErrMissingField("registry-user"),
		},
		{
			Name: "registry missing password",
			Options: &commands.CredentialSetOptions{
				ResourceOptions: testing.ValidResourceOptions,
				Registry:        "example.com",
				RegistryUser:    "projectriff",
			},
			// allow password to be blank
			ShouldValidate: true,
		},
		{
			Name: "multiple registries",
			Options: &commands.CredentialSetOptions{
				ResourceOptions:   testing.InvalidResourceOptions,
				DockerHubId:       "projectriff",
				DockerHubPassword: "1password",
				GcrTokenPath:      "gcr-credentials.json",
				Registry:          "example.com",
				RegistryUser:      "projectriff",
				RegistryPassword:  "1password",
			},
			ExpectFieldError: testing.InvalidResourceOptionsFieldError.Also(
				cli.ErrMultipleOneOf("docker-hub", "gcr", "registry"),
			),
		},

		{
			Name: "docker hub as default image prefix",
			Options: &commands.CredentialSetOptions{
				ResourceOptions:       testing.ValidResourceOptions,
				DockerHubId:           "projectriff",
				DockerHubPassword:     "1password",
				SetDefaultImagePrefix: true,
			},
			ShouldValidate: true,
		},
		{
			Name: "gcr as default image prefix",
			Options: &commands.CredentialSetOptions{
				ResourceOptions:       testing.ValidResourceOptions,
				GcrTokenPath:          "gcr-credentials.json",
				SetDefaultImagePrefix: true,
			},
			ShouldValidate: true,
		},
		{
			Name: "registry as default image prefix",
			Options: &commands.CredentialSetOptions{
				ResourceOptions:       testing.ValidResourceOptions,
				Registry:              "example.com",
				RegistryUser:          "projectriff",
				RegistryPassword:      "1password",
				SetDefaultImagePrefix: true,
			},
			ExpectFieldError: cli.ErrInvalidValue("cannot be used with registry", "set-default-image-prefix"),
		},
	}

	table.Run(t)
}

func TestCredentialSetCommand(t *testing.T) {
	credentialName := "test-credential"
	defaultNamespace := "default"
	credentialLabel := build.CredentialLabelKey
	dockerHubId := "projectriff"
	dockerHubPassword := "docker-password"
	registryHost := "https://example.com"
	registryUser := "projectriff"
	registryPassword := "registry-password"

	table := testing.CommandTable{
		{
			Name:        "invalid args",
			Args:        []string{},
			ShouldError: true,
		},
		{
			Name:  "create secret docker hub",
			Args:  []string{credentialName, "--docker-hub", dockerHubId},
			Stdin: []byte(dockerHubPassword),
			ExpectCreates: []runtime.Object{
				&corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      credentialName,
						Namespace: defaultNamespace,
						Labels:    map[string]string{credentialLabel: ""},
						Annotations: map[string]string{
							"build.knative.dev/docker-0": "https://index.docker.io/v1/",
						},
					},
					Type: corev1.SecretTypeBasicAuth,
					StringData: map[string]string{
						"username": dockerHubId,
						"password": dockerHubPassword,
					},
				},
			},
		},
		{
			Name: "create secret gcr",
			Args: []string{credentialName, "--gcr", "./testdata/gcr.json"},
			ExpectCreates: []runtime.Object{
				&corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      credentialName,
						Namespace: defaultNamespace,
						Labels:    map[string]string{credentialLabel: ""},
						Annotations: map[string]string{
							"build.knative.dev/docker-0": "https://gcr.io",
							"build.knative.dev/docker-1": "https://us.gcr.io",
							"build.knative.dev/docker-2": "https://eu.gcr.io",
							"build.knative.dev/docker-3": "https://asia.gcr.io",
						},
					},
					Type: corev1.SecretTypeBasicAuth,
					StringData: map[string]string{
						"username": "_json_key",
						"password": `{"project_id":"my-gcp-project"}`,
					},
				},
			},
		},
		{
			Name:        "create secret gcr, bad token path",
			Args:        []string{credentialName, "--gcr", "./testdata/gcr-badpath.json"},
			ShouldError: true,
		},
		{
			Name:        "create secret gcr, invalid token",
			Args:        []string{credentialName, "--gcr", "./testdata/gcr-invalid.json"},
			ShouldError: true,
		},
		{
			Name:  "create secret registry",
			Args:  []string{credentialName, "--registry", registryHost, "--registry-user", registryUser},
			Stdin: []byte(registryPassword),
			ExpectCreates: []runtime.Object{
				&corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      credentialName,
						Namespace: defaultNamespace,
						Labels:    map[string]string{credentialLabel: ""},
						Annotations: map[string]string{
							"build.knative.dev/docker-0": registryHost,
						},
					},
					Type: corev1.SecretTypeBasicAuth,
					StringData: map[string]string{
						"username": registryUser,
						"password": registryPassword,
					},
				},
			},
		},
		{
			Name:  "update secret",
			Args:  []string{credentialName, "--registry", registryHost, "--registry-user", registryUser},
			Stdin: []byte(registryPassword),
			GivenObjects: []runtime.Object{
				&corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      credentialName,
						Namespace: defaultNamespace,
						Labels:    map[string]string{credentialLabel: ""},
						Annotations: map[string]string{
							"build.knative.dev/docker-0": "https://index.dockerhub.io/projectriff",
						},
					},
					Type: corev1.SecretTypeBasicAuth,
					StringData: map[string]string{
						"username": dockerHubId,
						"password": dockerHubPassword,
					},
				},
			},
			ExpectUpdates: []runtime.Object{
				&corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      credentialName,
						Namespace: defaultNamespace,
						Labels:    map[string]string{credentialLabel: ""},
						Annotations: map[string]string{
							"build.knative.dev/docker-0": registryHost,
						},
					},
					Type: corev1.SecretTypeBasicAuth,
					StringData: map[string]string{
						"username": registryUser,
						"password": registryPassword,
					},
				},
			},
		},
		{
			Name:  "get error",
			Args:  []string{credentialName, "--registry", registryHost, "--registry-user", registryUser},
			Stdin: []byte(registryPassword),
			GivenObjects: []runtime.Object{
				&corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      credentialName,
						Namespace: defaultNamespace,
						Labels:    map[string]string{credentialLabel: ""},
						Annotations: map[string]string{
							"build.knative.dev/docker-0": "https://index.dockerhub.io/projectriff",
						},
					},
					Type: corev1.SecretTypeBasicAuth,
					StringData: map[string]string{
						"username": dockerHubId,
						"password": dockerHubPassword,
					},
				},
			},
			WithReactors: []testing.ReactionFunc{
				testing.InduceFailure("get", "secrets"),
			},
			ShouldError: true,
		},
		{
			Name:  "create error",
			Args:  []string{credentialName, "--registry", registryHost, "--registry-user", registryUser},
			Stdin: []byte(registryPassword),
			WithReactors: []testing.ReactionFunc{
				testing.InduceFailure("create", "secrets"),
			},
			ExpectCreates: []runtime.Object{
				&corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      credentialName,
						Namespace: defaultNamespace,
						Labels:    map[string]string{credentialLabel: ""},
						Annotations: map[string]string{
							"build.knative.dev/docker-0": registryHost,
						},
					},
					Type: corev1.SecretTypeBasicAuth,
					StringData: map[string]string{
						"username": registryUser,
						"password": registryPassword,
					},
				},
			},
			ShouldError: true,
		},
		{
			Name:  "update error",
			Args:  []string{credentialName, "--registry", registryHost, "--registry-user", registryUser},
			Stdin: []byte(registryPassword),
			GivenObjects: []runtime.Object{
				&corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      credentialName,
						Namespace: defaultNamespace,
						Labels:    map[string]string{credentialLabel: ""},
						Annotations: map[string]string{
							"build.knative.dev/docker-0": "https://index.dockerhub.io/projectriff",
						},
					},
					Type: corev1.SecretTypeBasicAuth,
					StringData: map[string]string{
						"username": dockerHubId,
						"password": dockerHubPassword,
					},
				},
			},
			WithReactors: []testing.ReactionFunc{
				testing.InduceFailure("update", "secrets"),
			},
			ExpectUpdates: []runtime.Object{
				&corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      credentialName,
						Namespace: defaultNamespace,
						Labels:    map[string]string{credentialLabel: ""},
						Annotations: map[string]string{
							"build.knative.dev/docker-0": registryHost,
						},
					},
					Type: corev1.SecretTypeBasicAuth,
					StringData: map[string]string{
						"username": registryUser,
						"password": registryPassword,
					},
				},
			},
			ShouldError: true,
		},
		{
			Name:  "no clobber",
			Args:  []string{"not-a-credential", "--registry", registryHost, "--registry-user", registryUser},
			Stdin: []byte(registryPassword),
			GivenObjects: []runtime.Object{
				&corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "not-a-credential",
						Namespace: defaultNamespace,
					},
					StringData: map[string]string{},
				},
			},
			ShouldError: true,
		},
		{
			Name:  "default image prefix create docker hub",
			Args:  []string{credentialName, "--docker-hub", dockerHubId, "--set-default-image-prefix"},
			Stdin: []byte(dockerHubPassword),
			ExpectCreates: []runtime.Object{
				&corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      credentialName,
						Namespace: defaultNamespace,
						Labels:    map[string]string{credentialLabel: ""},
						Annotations: map[string]string{
							"build.knative.dev/docker-0": "https://index.docker.io/v1/",
						},
					},
					Type: corev1.SecretTypeBasicAuth,
					StringData: map[string]string{
						"username": dockerHubId,
						"password": dockerHubPassword,
					},
				},
				&corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      "riff-build",
					},
					Data: map[string]string{
						"default-image-prefix": "docker.io/projectriff",
					},
				},
			},
		},
		{
			Name: "default image prefix create gcr",
			Args: []string{credentialName, "--gcr", "./testdata/gcr.json", "--set-default-image-prefix"},
			ExpectCreates: []runtime.Object{
				&corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      credentialName,
						Namespace: defaultNamespace,
						Labels:    map[string]string{credentialLabel: ""},
						Annotations: map[string]string{
							"build.knative.dev/docker-0": "https://gcr.io",
							"build.knative.dev/docker-1": "https://us.gcr.io",
							"build.knative.dev/docker-2": "https://eu.gcr.io",
							"build.knative.dev/docker-3": "https://asia.gcr.io",
						},
					},
					Type: corev1.SecretTypeBasicAuth,
					StringData: map[string]string{
						"username": "_json_key",
						"password": `{"project_id":"my-gcp-project"}`,
					},
				},
				&corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      "riff-build",
					},
					Data: map[string]string{
						"default-image-prefix": "gcr.io/my-gcp-project",
					},
				},
			},
		},
		{
			Name:        "default image prefix create registry",
			Args:        []string{credentialName, "--registry", registryHost, "--registry-user", registryUser, "--set-default-image-prefix"},
			Stdin:       []byte(registryPassword),
			ShouldError: true,
		},
		{
			Name:  "default image prefix update",
			Args:  []string{credentialName, "--docker-hub", dockerHubId, "--set-default-image-prefix"},
			Stdin: []byte(dockerHubPassword),
			GivenObjects: []runtime.Object{
				&corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      "riff-build",
					},
					Data: map[string]string{
						"existing-data":        "should still be here",
						"default-image-prefix": "gcr.io/my-gcp-project",
					},
				},
			},
			ExpectCreates: []runtime.Object{
				&corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      credentialName,
						Namespace: defaultNamespace,
						Labels:    map[string]string{credentialLabel: ""},
						Annotations: map[string]string{
							"build.knative.dev/docker-0": "https://index.docker.io/v1/",
						},
					},
					Type: corev1.SecretTypeBasicAuth,
					StringData: map[string]string{
						"username": dockerHubId,
						"password": dockerHubPassword,
					},
				},
			},
			ExpectUpdates: []runtime.Object{
				&corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      "riff-build",
					},
					Data: map[string]string{
						"existing-data":        "should still be here",
						"default-image-prefix": "docker.io/projectriff",
					},
				},
			},
		},
		{
			Name:  "default image prefix get error",
			Args:  []string{credentialName, "--docker-hub", dockerHubId, "--set-default-image-prefix"},
			Stdin: []byte(dockerHubPassword),
			WithReactors: []testing.ReactionFunc{
				testing.InduceFailure("get", "configmaps"),
			},
			ExpectCreates: []runtime.Object{
				&corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      credentialName,
						Namespace: defaultNamespace,
						Labels:    map[string]string{credentialLabel: ""},
						Annotations: map[string]string{
							"build.knative.dev/docker-0": "https://index.docker.io/v1/",
						},
					},
					Type: corev1.SecretTypeBasicAuth,
					StringData: map[string]string{
						"username": dockerHubId,
						"password": dockerHubPassword,
					},
				},
			},
			ShouldError: true,
		},
		{
			Name:  "default image prefix create error",
			Args:  []string{credentialName, "--docker-hub", dockerHubId, "--set-default-image-prefix"},
			Stdin: []byte(dockerHubPassword),
			WithReactors: []testing.ReactionFunc{
				testing.InduceFailure("create", "configmaps"),
			},
			ExpectCreates: []runtime.Object{
				&corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      credentialName,
						Namespace: defaultNamespace,
						Labels:    map[string]string{credentialLabel: ""},
						Annotations: map[string]string{
							"build.knative.dev/docker-0": "https://index.docker.io/v1/",
						},
					},
					Type: corev1.SecretTypeBasicAuth,
					StringData: map[string]string{
						"username": dockerHubId,
						"password": dockerHubPassword,
					},
				},
				&corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      "riff-build",
					},
					Data: map[string]string{
						"default-image-prefix": "docker.io/projectriff",
					},
				},
			},
			ShouldError: true,
		},
		{
			Name:  "default image prefix update",
			Args:  []string{credentialName, "--docker-hub", dockerHubId, "--set-default-image-prefix"},
			Stdin: []byte(dockerHubPassword),
			WithReactors: []testing.ReactionFunc{
				testing.InduceFailure("update", "configmaps"),
			},
			GivenObjects: []runtime.Object{
				&corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      "riff-build",
					},
					Data: map[string]string{
						"existing-data":        "should still be here",
						"default-image-prefix": "gcr.io/my-gcp-project",
					},
				},
			},
			ExpectCreates: []runtime.Object{
				&corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      credentialName,
						Namespace: defaultNamespace,
						Labels:    map[string]string{credentialLabel: ""},
						Annotations: map[string]string{
							"build.knative.dev/docker-0": "https://index.docker.io/v1/",
						},
					},
					Type: corev1.SecretTypeBasicAuth,
					StringData: map[string]string{
						"username": dockerHubId,
						"password": dockerHubPassword,
					},
				},
			},
			ExpectUpdates: []runtime.Object{
				&corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      "riff-build",
					},
					Data: map[string]string{
						"existing-data":        "should still be here",
						"default-image-prefix": "docker.io/projectriff",
					},
				},
			},
			ShouldError: true,
		},
	}

	table.Run(t, commands.NewCredentialSetCommand)
}
