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
	"testing"

	"github.com/projectriff/riff/cli/pkg/build/commands"
	"github.com/projectriff/riff/cli/pkg/cli"
	rifftesting "github.com/projectriff/riff/cli/pkg/testing"
	buildv1alpha1 "github.com/projectriff/riff/system/pkg/apis/build/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestCredentialApplyOptions(t *testing.T) {
	table := rifftesting.OptionsTable{
		{
			Name: "valid namespaced resource",
			Options: &commands.CredentialApplyOptions{
				ResourceOptions: rifftesting.ValidResourceOptions,
			},
			ExpectFieldErrors: cli.ErrMissingOneOf(cli.DockerHubFlagName, cli.GcrFlagName, cli.RegistryFlagName),
		},
		{
			Name: "invalid namespaced resource",
			Options: &commands.CredentialApplyOptions{
				ResourceOptions: rifftesting.InvalidResourceOptions,
			},
			ExpectFieldErrors: rifftesting.InvalidResourceOptionsFieldError.Also(
				cli.ErrMissingOneOf(cli.DockerHubFlagName, cli.GcrFlagName, cli.RegistryFlagName),
			),
		},
		{
			Name: "docker hub",
			Options: &commands.CredentialApplyOptions{
				ResourceOptions:   rifftesting.ValidResourceOptions,
				DockerHubId:       "projectriff",
				DockerHubPassword: []byte("1password"),
			},
			ShouldValidate: true,
		},
		{
			Name: "docker hub missing password",
			Options: &commands.CredentialApplyOptions{
				ResourceOptions: rifftesting.ValidResourceOptions,
				DockerHubId:     "projectriff",
			},
			ExpectFieldErrors: cli.ErrMissingField("<docker-hub-password>"),
		},
		{
			Name: "gcr",
			Options: &commands.CredentialApplyOptions{
				ResourceOptions: rifftesting.ValidResourceOptions,
				GcrTokenPath:    "gcr-credentials.json",
			},
			ShouldValidate: true,
		},
		{
			Name: "registry",
			Options: &commands.CredentialApplyOptions{
				ResourceOptions:  rifftesting.ValidResourceOptions,
				Registry:         "example.com",
				RegistryUser:     "projectriff",
				RegistryPassword: []byte("1password"),
			},
			ShouldValidate: true,
		},
		{
			Name: "registry missing user",
			Options: &commands.CredentialApplyOptions{
				ResourceOptions:  rifftesting.ValidResourceOptions,
				Registry:         "example.com",
				RegistryPassword: []byte("1password"),
			},
			ExpectFieldErrors: cli.ErrMissingField(cli.RegistryUserFlagName),
		},
		{
			Name: "registry missing password",
			Options: &commands.CredentialApplyOptions{
				ResourceOptions: rifftesting.ValidResourceOptions,
				Registry:        "example.com",
				RegistryUser:    "projectriff",
			},
			// allow password to be blank
			ShouldValidate: true,
		},
		{
			Name: "multiple registries",
			Options: &commands.CredentialApplyOptions{
				ResourceOptions:   rifftesting.InvalidResourceOptions,
				DockerHubId:       "projectriff",
				DockerHubPassword: []byte("1password"),
				GcrTokenPath:      "gcr-credentials.json",
				Registry:          "example.com",
				RegistryUser:      "projectriff",
				RegistryPassword:  []byte("1password"),
			},
			ExpectFieldErrors: rifftesting.InvalidResourceOptionsFieldError.Also(
				cli.ErrMultipleOneOf(cli.DockerHubFlagName, cli.GcrFlagName, cli.RegistryFlagName),
			),
		},
		{
			Name: "docker hub as default image prefix",
			Options: &commands.CredentialApplyOptions{
				ResourceOptions:       rifftesting.ValidResourceOptions,
				DockerHubId:           "projectriff",
				DockerHubPassword:     []byte("1password"),
				SetDefaultImagePrefix: true,
			},
			ShouldValidate: true,
		},
		{
			Name: "gcr as default image prefix",
			Options: &commands.CredentialApplyOptions{
				ResourceOptions:       rifftesting.ValidResourceOptions,
				GcrTokenPath:          "gcr-credentials.json",
				SetDefaultImagePrefix: true,
			},
			ShouldValidate: true,
		},
		{
			Name: "registry as default image prefix",
			Options: &commands.CredentialApplyOptions{
				ResourceOptions:       rifftesting.ValidResourceOptions,
				Registry:              "example.com",
				RegistryUser:          "projectriff",
				RegistryPassword:      []byte("1password"),
				SetDefaultImagePrefix: true,
			},
			ExpectFieldErrors: cli.ErrInvalidValue(fmt.Sprintf("cannot be used with %s, without %s", cli.RegistryFlagName, cli.DefaultImagePrefixFlagName), cli.SetDefaultImagePrefixFlagName),
		},
		{
			Name: "dry run",
			Options: &commands.CredentialApplyOptions{
				ResourceOptions:   rifftesting.ValidResourceOptions,
				DockerHubId:       "projectriff",
				DockerHubPassword: []byte("1password"),
				DryRun:            true,
			},
			ShouldValidate: true,
		},
	}

	table.Run(t)
}

func TestCredentialApplyCommand(t *testing.T) {
	credentialName := "test-credential"
	defaultNamespace := "default"
	credentialLabel := buildv1alpha1.CredentialLabelKey
	dockerHubId := "projectriff"
	dockerHubPassword := "docker-password"
	registryHost := "example.com"
	registryURL := fmt.Sprintf("https://%s", registryHost)
	registryUser := "projectriff"
	registryPassword := "registry-password"

	table := rifftesting.CommandTable{
		{
			Name:        "invalid args",
			Args:        []string{},
			ShouldError: true,
		},
		{
			Name:  "create secret docker hub",
			Args:  []string{credentialName, cli.DockerHubFlagName, dockerHubId},
			Stdin: []byte(dockerHubPassword),
			ExpectCreates: []runtime.Object{
				&corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      credentialName,
						Namespace: defaultNamespace,
						Labels:    map[string]string{credentialLabel: "docker-hub"},
						Annotations: map[string]string{
							"build.knative.dev/docker-0": "https://index.docker.io/v1/",
							"build.pivotal.io/docker":    "https://index.docker.io/v1/",
						},
					},
					Type: corev1.SecretTypeBasicAuth,
					StringData: map[string]string{
						"username": dockerHubId,
						"password": dockerHubPassword,
					},
				},
			},
			ExpectOutput: `
Apply credentials "test-credential"
`,
		},
		{
			Name: "create secret gcr",
			Args: []string{credentialName, cli.GcrFlagName, "./testdata/gcr.json"},
			ExpectCreates: []runtime.Object{
				&corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      credentialName,
						Namespace: defaultNamespace,
						Labels:    map[string]string{credentialLabel: "gcr"},
						Annotations: map[string]string{
							"build.knative.dev/docker-0": "https://gcr.io",
							"build.knative.dev/docker-1": "https://us.gcr.io",
							"build.knative.dev/docker-2": "https://eu.gcr.io",
							"build.knative.dev/docker-3": "https://asia.gcr.io",
							"build.pivotal.io/docker":    "https://gcr.io",
						},
					},
					Type: corev1.SecretTypeBasicAuth,
					StringData: map[string]string{
						"username": "_json_key",
						"password": `{"project_id":"my-gcp-project"}`,
					},
				},
			},
			ExpectOutput: `
Apply credentials "test-credential"
`,
		},
		{
			Name:        "create secret gcr, bad token path",
			Args:        []string{credentialName, cli.GcrFlagName, "./testdata/gcr-badpath.json"},
			ShouldError: true,
		},
		{
			Name:        "create secret gcr, invalid token",
			Args:        []string{credentialName, cli.GcrFlagName, "./testdata/gcr-invalid.json"},
			ShouldError: true,
		},
		{
			Name:  "create secret registry",
			Args:  []string{credentialName, cli.RegistryFlagName, registryURL, cli.RegistryUserFlagName, registryUser},
			Stdin: []byte(registryPassword),
			ExpectCreates: []runtime.Object{
				&corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      credentialName,
						Namespace: defaultNamespace,
						Labels:    map[string]string{credentialLabel: "basic-auth"},
						Annotations: map[string]string{
							"build.knative.dev/docker-0": registryURL,
							"build.pivotal.io/docker":    registryURL,
						},
					},
					Type: corev1.SecretTypeBasicAuth,
					StringData: map[string]string{
						"username": registryUser,
						"password": registryPassword,
					},
				},
			},
			ExpectOutput: `
Apply credentials "test-credential"
`,
		},
		{
			Name:  "update secret",
			Args:  []string{credentialName, cli.RegistryFlagName, registryURL, cli.RegistryUserFlagName, registryUser},
			Stdin: []byte(registryPassword),
			GivenObjects: []runtime.Object{
				&corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      credentialName,
						Namespace: defaultNamespace,
						Labels:    map[string]string{credentialLabel: "docker-hub"},
						Annotations: map[string]string{
							"build.knative.dev/docker-0": "https://index.docker.io/v1/",
							"build.pivotal.io/docker":    "https://index.docker.io/v1/",
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
						Labels:    map[string]string{credentialLabel: "basic-auth"},
						Annotations: map[string]string{
							"build.knative.dev/docker-0": registryURL,
							"build.pivotal.io/docker":    registryURL,
						},
					},
					Type: corev1.SecretTypeBasicAuth,
					StringData: map[string]string{
						"username": registryUser,
						"password": registryPassword,
					},
				},
			},
			ExpectOutput: `
Apply credentials "test-credential"
`,
		},
		{
			Name:  "get error",
			Args:  []string{credentialName, cli.RegistryFlagName, registryURL, cli.RegistryUserFlagName, registryUser},
			Stdin: []byte(registryPassword),
			GivenObjects: []runtime.Object{
				&corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      credentialName,
						Namespace: defaultNamespace,
						Labels:    map[string]string{credentialLabel: "docker-hub"},
						Annotations: map[string]string{
							"build.knative.dev/docker-0": "https://index.docker.io/v1/",
							"build.pivotal.io/docker":    "https://index.docker.io/v1/",
						},
					},
					Type: corev1.SecretTypeBasicAuth,
					StringData: map[string]string{
						"username": dockerHubId,
						"password": dockerHubPassword,
					},
				},
			},
			WithReactors: []rifftesting.ReactionFunc{
				rifftesting.InduceFailure("get", "secrets"),
			},
			ShouldError: true,
		},
		{
			Name:  "create error",
			Args:  []string{credentialName, cli.RegistryFlagName, registryURL, cli.RegistryUserFlagName, registryUser},
			Stdin: []byte(registryPassword),
			WithReactors: []rifftesting.ReactionFunc{
				rifftesting.InduceFailure("create", "secrets"),
			},
			ExpectCreates: []runtime.Object{
				&corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      credentialName,
						Namespace: defaultNamespace,
						Labels:    map[string]string{credentialLabel: "basic-auth"},
						Annotations: map[string]string{
							"build.knative.dev/docker-0": registryURL,
							"build.pivotal.io/docker":    registryURL,
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
			Args:  []string{credentialName, cli.RegistryFlagName, registryURL, cli.RegistryUserFlagName, registryUser},
			Stdin: []byte(registryPassword),
			GivenObjects: []runtime.Object{
				&corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      credentialName,
						Namespace: defaultNamespace,
						Labels:    map[string]string{credentialLabel: "docker-hub"},
						Annotations: map[string]string{
							"build.knative.dev/docker-0": "https://index.docker.io/v1/",
							"build.pivotal.io/docker":    "https://index.docker.io/v1/",
						},
					},
					Type: corev1.SecretTypeBasicAuth,
					StringData: map[string]string{
						"username": dockerHubId,
						"password": dockerHubPassword,
					},
				},
			},
			WithReactors: []rifftesting.ReactionFunc{
				rifftesting.InduceFailure("update", "secrets"),
			},
			ExpectUpdates: []runtime.Object{
				&corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      credentialName,
						Namespace: defaultNamespace,
						Labels:    map[string]string{credentialLabel: "basic-auth"},
						Annotations: map[string]string{
							"build.knative.dev/docker-0": registryURL,
							"build.pivotal.io/docker":    registryURL,
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
			Args:  []string{"not-a-credential", cli.RegistryFlagName, registryURL, cli.RegistryUserFlagName, registryUser},
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
			Args:  []string{credentialName, cli.DockerHubFlagName, dockerHubId, cli.SetDefaultImagePrefixFlagName},
			Stdin: []byte(dockerHubPassword),
			ExpectCreates: []runtime.Object{
				&corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      credentialName,
						Namespace: defaultNamespace,
						Labels:    map[string]string{credentialLabel: "docker-hub"},
						Annotations: map[string]string{
							"build.knative.dev/docker-0": "https://index.docker.io/v1/",
							"build.pivotal.io/docker":    "https://index.docker.io/v1/",
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
			ExpectOutput: `
Apply credentials "test-credential"
Set default image prefix to "docker.io/projectriff"
`,
		},
		{
			Name: "default image prefix create gcr",
			Args: []string{credentialName, cli.GcrFlagName, "./testdata/gcr.json", cli.SetDefaultImagePrefixFlagName},
			ExpectCreates: []runtime.Object{
				&corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      credentialName,
						Namespace: defaultNamespace,
						Labels:    map[string]string{credentialLabel: "gcr"},
						Annotations: map[string]string{
							"build.knative.dev/docker-0": "https://gcr.io",
							"build.knative.dev/docker-1": "https://us.gcr.io",
							"build.knative.dev/docker-2": "https://eu.gcr.io",
							"build.knative.dev/docker-3": "https://asia.gcr.io",
							"build.pivotal.io/docker":    "https://gcr.io",
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
			ExpectOutput: `
Apply credentials "test-credential"
Set default image prefix to "gcr.io/my-gcp-project"
`,
		},
		{
			Name:  "default image prefix create registry, explicit default",
			Args:  []string{credentialName, cli.RegistryFlagName, registryURL, cli.RegistryUserFlagName, registryUser, cli.DefaultImagePrefixFlagName, registryHost},
			Stdin: []byte(registryPassword),
			ExpectCreates: []runtime.Object{
				&corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      credentialName,
						Namespace: defaultNamespace,
						Labels:    map[string]string{credentialLabel: "basic-auth"},
						Annotations: map[string]string{
							"build.knative.dev/docker-0": registryURL,
							"build.pivotal.io/docker":    registryURL,
						},
					},
					Type: corev1.SecretTypeBasicAuth,
					StringData: map[string]string{
						"username": registryUser,
						"password": registryPassword,
					},
				},
				&corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      "riff-build",
					},
					Data: map[string]string{
						"default-image-prefix": registryHost,
					},
				},
			},
			ExpectOutput: `
Apply credentials "test-credential"
Set default image prefix to "example.com"
`,
		},
		{
			Name:        "default image prefix create registry, no implicit default",
			Args:        []string{credentialName, cli.RegistryFlagName, registryURL, cli.RegistryUserFlagName, registryUser, cli.SetDefaultImagePrefixFlagName},
			Stdin:       []byte(registryPassword),
			ShouldError: true,
		},
		{
			Name:  "default image prefix update",
			Args:  []string{credentialName, cli.DockerHubFlagName, dockerHubId, cli.SetDefaultImagePrefixFlagName},
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
						Labels:    map[string]string{credentialLabel: "docker-hub"},
						Annotations: map[string]string{
							"build.knative.dev/docker-0": "https://index.docker.io/v1/",
							"build.pivotal.io/docker":    "https://index.docker.io/v1/",
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
			ExpectOutput: `
Apply credentials "test-credential"
Set default image prefix to "docker.io/projectriff"
`,
		},
		{
			Name:  "default image prefix get error",
			Args:  []string{credentialName, cli.DockerHubFlagName, dockerHubId, cli.SetDefaultImagePrefixFlagName},
			Stdin: []byte(dockerHubPassword),
			WithReactors: []rifftesting.ReactionFunc{
				rifftesting.InduceFailure("get", "configmaps"),
			},
			ExpectCreates: []runtime.Object{
				&corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      credentialName,
						Namespace: defaultNamespace,
						Labels:    map[string]string{credentialLabel: "docker-hub"},
						Annotations: map[string]string{
							"build.knative.dev/docker-0": "https://index.docker.io/v1/",
							"build.pivotal.io/docker":    "https://index.docker.io/v1/",
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
			Args:  []string{credentialName, cli.DockerHubFlagName, dockerHubId, cli.SetDefaultImagePrefixFlagName},
			Stdin: []byte(dockerHubPassword),
			WithReactors: []rifftesting.ReactionFunc{
				rifftesting.InduceFailure("create", "configmaps"),
			},
			ExpectCreates: []runtime.Object{
				&corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      credentialName,
						Namespace: defaultNamespace,
						Labels:    map[string]string{credentialLabel: "docker-hub"},
						Annotations: map[string]string{
							"build.knative.dev/docker-0": "https://index.docker.io/v1/",
							"build.pivotal.io/docker":    "https://index.docker.io/v1/",
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
			Args:  []string{credentialName, cli.DockerHubFlagName, dockerHubId, cli.SetDefaultImagePrefixFlagName},
			Stdin: []byte(dockerHubPassword),
			WithReactors: []rifftesting.ReactionFunc{
				rifftesting.InduceFailure("update", "configmaps"),
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
						Labels:    map[string]string{credentialLabel: "docker-hub"},
						Annotations: map[string]string{
							"build.knative.dev/docker-0": "https://index.docker.io/v1/",
							"build.pivotal.io/docker":    "https://index.docker.io/v1/",
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
		{
			Name:  "create dry run",
			Args:  []string{credentialName, cli.DockerHubFlagName, dockerHubId, cli.SetDefaultImagePrefixFlagName, cli.DryRunFlagName},
			Stdin: []byte(dockerHubPassword),
			ExpectOutput: `
---
apiVersion: v1
kind: Secret
metadata:
  annotations:
    build.knative.dev/docker-0: https://index.docker.io/v1/
    build.pivotal.io/docker: https://index.docker.io/v1/
  creationTimestamp: null
  labels:
    build.projectriff.io/credential: docker-hub
  name: test-credential
  namespace: default
stringData:
  password: docker-password
  username: projectriff
type: kubernetes.io/basic-auth

Apply credentials "test-credential"
---
apiVersion: v1
data:
  default-image-prefix: docker.io/projectriff
kind: ConfigMap
metadata:
  creationTimestamp: null
  name: riff-build
  namespace: default

Set default image prefix to "docker.io/projectriff"
`,
		},
		{
			Name:  "update dry run",
			Args:  []string{credentialName, cli.DockerHubFlagName, dockerHubId, cli.SetDefaultImagePrefixFlagName, cli.DryRunFlagName},
			Stdin: []byte(dockerHubPassword),
			GivenObjects: []runtime.Object{
				&corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      credentialName,
						Namespace: defaultNamespace,
						Labels:    map[string]string{credentialLabel: ""},
					},
				},
				&corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: defaultNamespace,
						Name:      "riff-build",
					},
					Data: map[string]string{},
				},
			},
			ExpectOutput: `
---
apiVersion: v1
kind: Secret
metadata:
  annotations:
    build.knative.dev/docker-0: https://index.docker.io/v1/
    build.pivotal.io/docker: https://index.docker.io/v1/
  creationTimestamp: null
  labels:
    build.projectriff.io/credential: docker-hub
  name: test-credential
  namespace: default
stringData:
  password: docker-password
  username: projectriff
type: kubernetes.io/basic-auth

Apply credentials "test-credential"
---
apiVersion: v1
data:
  default-image-prefix: docker.io/projectriff
kind: ConfigMap
metadata:
  creationTimestamp: null
  name: riff-build
  namespace: default

Set default image prefix to "docker.io/projectriff"
`,
		},
	}

	table.Run(t, commands.NewCredentialApplyCommand)
}
