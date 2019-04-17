/*
 * Copyright 2018-2019 The original author or authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     https://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package commands

import (
	"fmt"
	"sort"
	"strings"

	"github.com/projectriff/riff/pkg/core"
	"github.com/projectriff/riff/pkg/env"
	"github.com/spf13/cobra"
)

func Namespace() *cobra.Command {
	return &cobra.Command{
		Use:   "namespace",
		Short: fmt.Sprintf("Manage namespaces used for %s resources", env.Cli.Name),
	}
}

const (
	namespaceInitNameIndex = iota
	namespaceInitNumberOfArgs
)

const (
	namespaceCleanupNameIndex = iota
	namespaceCleanupNumberOfArgs
)

func NamespaceInit(manifests map[string]*core.Manifest, c *core.Client) *cobra.Command {
	options := core.NamespaceInitOptions{}

	var namedManifests []string
	for k, _ := range manifests {
		namedManifests = append(namedManifests, k)
	}
	sort.Strings(namedManifests)

	command := &cobra.Command{
		Use:     "init",
		Short:   "initialize " + env.Cli.Name + " resources in the namespace",
		Example: `  ` + env.Cli.Name + ` namespace init default --secret build-secret`,
		Args: ArgValidationConjunction(
			cobra.ExactArgs(namespaceInitNumberOfArgs),
			AtPosition(namespaceInitNameIndex, ValidName()),
		),
		PreRunE: FlagsValidatorAsCobraRunE(
			FlagsValidationConjunction(
				AtMostOneOf("gcr", "docker-hub", "dockerhub", "no-secret", "registry-user"),
				AtMostOneOf("secret", "no-secret"),
				NotBlank("secret"),
				FlagsDependency(Set("registry-user"), NotBlank("registry")),
				// NOTE this should be relaxed when we add support for bearer auth
				FlagsDependency(Set("registry"), NotBlank("registry-user")),
				FlagsDependency(Set("registry"), SupportedRegistryProtocol(func() string {
					return options.Registry
				})),
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			nsName := args[namespaceInitNameIndex]
			options.NamespaceName = nsName

			if options.Registry != "" && !strings.Contains(options.Registry, "://") {
				// default to https
				options.Registry = fmt.Sprintf("https://%s", options.Registry)
			}

			err := (*c).NamespaceInit(manifests, options)
			if err != nil {
				return err
			}

			PrintSuccessfulCompletion(cmd)
			return nil
		},
	}

	LabelArgs(command, "NAME")

	if len(namedManifests) > 0 {
		desc := fmt.Sprintf("manifest of kubernetes configuration files to be applied; can be a named manifest (%s) or a path of a manifest file", strings.Join(namedManifests, ", "))
		command.Flags().StringVarP(&options.Manifest, "manifest", "m", "stable", desc)
	} else {
		command.Flags().StringVarP(&options.Manifest, "manifest", "m", "", "path to a manifest of kubernetes configuration files to be applied")
		_ = command.MarkFlagRequired("manifest")
	}

	command.Flags().BoolVarP(&options.NoSecret, "no-secret", "", false, "no secret required for the image registry")
	command.Flags().StringVarP(&options.SecretName, "secret", "s", "push-credentials", "the name of a `secret` containing credentials for the image registry")
	command.Flags().StringVar(&options.GcrTokenPath, "gcr", "", "path to a file containing Google Container Registry credentials")
	// ToDo: remove this deprecated flag in favor of --docker-hub
	command.Flags().StringVar(&options.DockerHubId, "dockerhub", "", "dockerhub username for authentication; password will be read from stdin")
	command.Flags().MarkHidden("dockerhub")
	command.Flags().MarkDeprecated("dockerhub", "use --docker-hub instead")
	command.Flags().StringVar(&options.DockerHubId, "docker-hub", "", "Docker ID for authenticating with Docker Hub; password will be read from stdin")
	command.Flags().StringVar(&options.Registry, "registry", "", "registry server host, scheme must be \"http\" or \"https\" (default \"https\")")
	command.Flags().StringVar(&options.RegistryUser, "registry-user", "", "registry username; password will be read from stdin")
	command.Flags().StringVar(&options.ImagePrefix, "image-prefix", "", "image prefix to use for commands that would otherwise require an --image argument. If not set, this value will be derived for Docker Hub and GCR")

	return command
}

func NamespaceCleanup(c *core.Client) *cobra.Command {
	options := core.NamespaceCleanupOptions{}

	command := &cobra.Command{
		Use:   "cleanup",
		Short: "cleans up " + env.Cli.Name + " resources in the namespace",
		Long:  fmt.Sprintf(`cleans up %s resources in the namespace and the namespace itself if "--remove-ns" is set`, env.Cli.Name),
		Example: fmt.Sprintf(`  %[1]s namespace cleanup my-ns
  %[1]s namespace cleanup my-ns --remove-ns`, env.Cli.Name),
		Args: ArgValidationConjunction(
			cobra.ExactArgs(namespaceCleanupNumberOfArgs),
			AtPosition(namespaceCleanupNameIndex, ValidName()),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			options.NamespaceName = args[namespaceCleanupNameIndex]
			if options.NamespaceName == "default" && options.RemoveNamespace {
				return fmt.Errorf("cleanup canceled: the default namespace cannot be removed")
			}
			if err := (*c).NamespaceCleanup(options); err != nil {
				return err
			}

			PrintSuccessfulCompletion(cmd)
			return nil
		},
	}

	LabelArgs(command, "NAMESPACE")

	command.Flags().BoolVarP(&options.RemoveNamespace, "remove-ns", "", false, "removes the (non-default) namespace as well")

	return command
}

func SupportedRegistryProtocol(getRegistry func() string) func(cmd *cobra.Command) error {
	return func(cmd *cobra.Command) error {
		registry := getRegistry()
		if parts := strings.SplitN(registry, "://", 2); len(parts) == 2 && parts[0] != "http" && parts[0] != "https" {
			return fmt.Errorf("valid protocols are: %q, %q, found: %q", "http", "https", parts[0])
		}
		return nil
	}
}
