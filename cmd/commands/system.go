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

func SystemInstall(manifests map[string]*core.Manifest, kc *core.KubectlClient) *cobra.Command {
	options := core.SystemInstallOptions{}

	var namedManifests []string
	for k, _ := range manifests {
		namedManifests = append(namedManifests, k)
	}
	sort.Strings(namedManifests)

	command := &cobra.Command{
		Use:   "install",
		Short: "Install " + env.Cli.Name + " and Knative system components",
		Long: "Install " + env.Cli.Name + " and Knative system components.\n" +
			"\nIf an `istio-system` namespace isn't found, it will be created and Istio components will be installed. " +
			"\nUse the `--node-port` flag when installing on Minikube and other clusters that don't support an external load balancer. " +
			"\nUse the `--manifest` flag to specify the path or URL of a manifest file which provides the URLs of the kubernetes configuration files of the " +
			"components to be installed. The manifest file contents should be of the following form:" +
			`

    manifestVersion: 0.1
    istio:
    - https://path/to/istio-release.yaml
    knative:
    - https://path/to/build-release.yaml
    - https://path/to/serving-release.yaml
    - https://path/to/eventing-release.yaml
    namespace:
    - https://path/to/buildtemplate-release.yaml
` +
			"\nNote: relative file paths or http/https URLs may be used in the manifest.",
		Args: cobra.ExactArgs(systemInstallNumberOfArgs),
		RunE: func(cmd *cobra.Command, args []string) error {
			complete, err := (*kc).SystemInstall(manifests, options)
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

	if len(namedManifests) > 0 {
		desc := fmt.Sprintf("manifest of kubernetes configuration files to be applied; can be a named manifest (%s) or a path of a manifest file", strings.Join(namedManifests, ", "))
		command.Flags().StringVarP(&options.Manifest, "manifest", "m", "stable", desc)
	} else {
		command.Flags().StringVarP(&options.Manifest, "manifest", "m", "", "path to a manifest of kubernetes configuration files to be applied")
		command.MarkFlagRequired("manifest")
	}

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
