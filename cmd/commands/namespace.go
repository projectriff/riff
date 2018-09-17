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
	"github.com/projectriff/riff/pkg/core"
	"github.com/spf13/cobra"
)

func Namespace() *cobra.Command {
	return &cobra.Command{
		Use:   "namespace",
		Short: "Manage namespaces used for riff resources",
	}
}

const (
	namespaceInitNameIndex = iota
	namespaceInitNumberOfArgs
)

func NamespaceInit(kc *core.KubectlClient) *cobra.Command {
	options := core.NamespaceInitOptions{}

	command := &cobra.Command{
		Use:     "init",
		Short:   "initialize riff resources in the namespace",
		Example: `  riff namespace init default --secret build-secret`,
		Args: ArgValidationConjunction(
			cobra.ExactArgs(namespaceInitNumberOfArgs),
			AtPosition(namespaceInitNameIndex, ValidName()),
		),
		PreRunE: FlagsValidatorAsCobraRunE(
			FlagsValidationConjunction(
				AtMostOneOf("gcr", "dockerhub"),
				NotBlank("secret"),
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			nsName := args[channelCreateNameIndex]
			options.NamespaceName = nsName
			err := (*kc).NamespaceInit(options)
			if err != nil {
				return err
			}

			printSuccessfulCompletion(cmd)
			return nil
		},
	}

	LabelArgs(command, "NAME")

	command.Flags().StringVarP(&options.Manifest, "manifest", "m", "stable", "manifest of YAML files to be applied; can be a named manifest (stable or latest) or a path or URL of a manifest file")

	command.Flags().StringVarP(&options.SecretName, "secret", "s", "push-credentials", "the name of a `secret` containing credentials for the image registry")
	command.Flags().StringVar(&options.GcrTokenPath, "gcr", "", "path to a file containing Google Container Registry credentials")
	command.Flags().StringVar(&options.DockerHubUsername, "dockerhub", "", "dockerhub username for authentication; password will be read from stdin")

	return command
}
