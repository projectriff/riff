/*
 * Copyright 2018 The original author or authors
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
	"github.com/projectriff/riff/pkg/core"
	"github.com/projectriff/riff/pkg/env"
	"github.com/spf13/cobra"
)

const (
	systemInstallNumberOfArgs = iota
)

const (
	systemUninstallNumberOfArgs = iota
)

func System() *cobra.Command {
	return &cobra.Command{
		Use:   "system",
		Short: "Manage system related resources",
	}
}

func SystemInstall(manifests map[string]*core.Manifest, c *core.Client) *cobra.Command {
	options := core.SystemInstallOptions{}

	command := &cobra.Command{
		Use:   "install",
		Short: "Install " + env.Cli.Name + " and Knative system components",
		Long: "Install " + env.Cli.Name + " and Knative system components.\n" +
			"\nIf an `istio-system` namespace isn't found, it will be created and Istio components will be installed. " +
			"\nUse the `--node-port` flag when installing on Minikube and other clusters that don't support an external load balancer. " +
			"\nUse the `--manifest` flag to specify the path or URL of a manifest file which provides the URLs of the kubernetes configuration files of the " +
			"components to be installed. The manifest file contents should be of the following form:" +
			`
    kind: RiffSystem
    apiVersion: projectriff.io/v1alpha1
    metadata:
      name: riff-install
      creationTimestamp:
      labels:
        riff-install: 'true'
    spec:
      resources:
      - path: https://storage.googleapis.com/knative-releases/serving/previous/v0.2.2/istio.yaml
        name: istio
        checks:
        - kind: Pod
          namespace: istio-system
          selector:
            matchLabels:
              istio: citadel
          jsonPath: ".status.phase"
          pattern: Running
    status: {}
` +
			"\nNote: relative file paths or http/https URLs may be used in the manifest.",
		Args: cobra.ExactArgs(systemInstallNumberOfArgs),
		RunE: func(cmd *cobra.Command, args []string) error {
			complete, err := (*c).SystemInstall(manifests, options)
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

	command.Flags().StringVarP(&options.Manifest, "manifest", "m", "stable", "manifest of kubernetes configuration files to be applied; can be a named manifest (stable or latest) or a path of a manifest file")
	command.Flags().BoolVarP(&options.NodePort, "node-port", "", false, "whether to use NodePort instead of LoadBalancer for ingress gateways")
	command.Flags().BoolVarP(&options.Force, "force", "", false, "force the install of components without getting any prompts")

	return command
}

func SystemUninstall(kc *core.KubectlClient) *cobra.Command {
	options := core.SystemUninstallOptions{}

	command := &cobra.Command{
		Use:     "uninstall",
		Short:   "Remove " + env.Cli.Name + " and Knative system components",
		Long:    "Remove " + env.Cli.Name + " and Knative system components.\n\nUse the `--istio` flag to also remove Istio components.",
		Example: `  ` + env.Cli.Name + ` system uninstall`,
		Args:    cobra.ExactArgs(systemUninstallNumberOfArgs),
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
