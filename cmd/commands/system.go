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
	"errors"

	"github.com/projectriff/riff/pkg/core"
	"github.com/spf13/cobra"
)

func System() *cobra.Command {
	return &cobra.Command{
		Use:   "system",
		Short: "Manage system related resources",
	}
}

func SystemInstall(kc *core.KubectlClient) *cobra.Command {
	options := core.SystemInstallOptions{}

	command := &cobra.Command{
		Use:   "install",
		Short: "Install riff and Knative system components",
		Long: "Install riff and Knative system components.\n" +
			"\nIf an `istio-system` namespace isn't found, it will be created and Istio components will be installed. " +
			"\nUse the `--node-port` flag when installing on Minikube and other clusters that don't support an external load balancer. " +
			"\nUse the `--manifest` flag to specify the path of a manifest file which provides the URLs of the YAML definitions of the " +
			"components to be installed. The manifest file contents should be of the following form:" +
			`

    manifestVersion: 0.1
    istio:
    - https://path/to/istio-release.yaml
    knative:
    - https://path/to/serving-release.yaml
    - https://path/to/eventing-release.yaml
    - https://path/to/stub-bus-release.yaml
    namespace:
    - https://path/to/riff-buildtemplate-release.yaml
` +
			"\nNote: relative file paths or http/https URLs may be used in the manifest.",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			// TODO: implement support for global flags - for now don't allow their use
			if cmd.Flags().Changed("kubeconfig") {
				return errors.New("The 'kubeconfig' flag is not yet supported by the 'system install' command")
			}
			m, _ := cmd.Flags().GetString("master")
			if len(m) > 0 {
				return errors.New("The 'master' flag is not yet supported by the 'system install' command")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			complete, err := (*kc).SystemInstall(options)
			if err != nil {
				return err
			}

			if complete {
				PrintSuccessfulCompletion(cmd)
			} else {
				PrintInterruptedCompletion(cmd)
			}
			return nil
		},
	}

	command.Flags().StringVarP(&options.Manifest, "manifest", "m", "stable", "manifest of YAML files to be applied; can be a named manifest (stable or latest) or a file path of a manifest file")
	command.Flags().BoolVarP(&options.NodePort, "node-port", "", false, "whether to use NodePort instead of LoadBalancer for ingress gateways")
	command.Flags().BoolVarP(&options.Force, "force", "", false, "force the install of components without getting any prompts")

	return command
}

func SystemUninstall(kc *core.KubectlClient) *cobra.Command {
	options := core.SystemUninstallOptions{}

	command := &cobra.Command{
		Use:     "uninstall",
		Short:   "Remove riff and Knative system components",
		Long:    "Remove riff and Knative system components.\n\nUse the `--istio` flag to also remove Istio components.",
		Example: `  riff system uninstall`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			// TODO: implement support for global flags - for now don't allow their use
			if cmd.Flags().Changed("kubeconfig") {
				return errors.New("The 'kubeconfig' flag is not yet supported by the 'system install' command")
			}
			m, _ := cmd.Flags().GetString("master")
			if len(m) > 0 {
				return errors.New("The 'master' flag is not yet supported by the 'system install' command")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			complete, err := (*kc).SystemUninstall(options)
			if err != nil {
				return err
			}
			if complete {
				PrintSuccessfulCompletion(cmd)
			} else {
				PrintInterruptedCompletion(cmd)
			}
			return nil
		},
	}

	command.Flags().BoolVarP(&options.Istio, "istio", "", false, "include Istio and the istio-system namespace in the removal")
	command.Flags().BoolVarP(&options.Force, "force", "", false, "force the removal of components without getting any prompts")

	return command
}
