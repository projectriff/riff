/*
 * Copyright 2018 The original author or authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
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

func NamespaceInit(manifests map[string]*core.Manifest, kc *core.KubectlClient) *cobra.Command {
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
				AtMostOneOf("gcr", "dockerhub", "no-secret"),
				AtMostOneOf("secret", "no-secret"),
				NotBlank("secret"),
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			nsName := args[channelCreateNameIndex]
			options.NamespaceName = nsName
			err := (*kc).NamespaceInit(manifests, options)
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
		command.MarkFlagRequired("manifest")
	}

	command.Flags().BoolVarP(&options.NoSecret, "no-secret", "", false, "no secret required for the image registry")
	command.Flags().StringVarP(&options.SecretName, "secret", "s", "push-credentials", "the name of a `secret` containing credentials for the image registry")
	command.Flags().StringVar(&options.GcrTokenPath, "gcr", "", "path to a file containing Google Container Registry credentials")
	command.Flags().StringVar(&options.DockerHubUsername, "dockerhub", "", "dockerhub username for authentication; password will be read from stdin")

	return command
}
