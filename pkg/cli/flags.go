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

package cli

import (
	"strings"

	"github.com/spf13/cobra"
)

const (
	AllFlagName                   = "--all"
	AllNamespacesFlagName         = "--all-namespaces"
	ApplicationRefFlagName        = "--application-ref"
	ArtifactFlagName              = "--artifact"
	CacheSizeFlagName             = "--cache-size"
	ConfigFlagName                = "--config"
	DirectoryFlagName             = "--directory"
	DockerHubFlagName             = "--docker-hub"
	EnvFlagName                   = "--env"
	FunctionRefFlagName           = "--function-ref"
	GcrFlagName                   = "--gcr"
	GitRepoFlagName               = "--git-repo"
	GitRevisionFlagName           = "--git-revision"
	HandlerFlagName               = "--handler"
	ImageFlagName                 = "--image"
	InputFlagName                 = "--input"
	InvokerFlagName               = "--invoker"
	JSONFlagName                  = "--json"
	KubeConfigFlagName            = "--kube-config"
	LocalPathFlagName             = "--local-path"
	NamespaceFlagName             = "--namespace"
	NoColorFlagName               = "--no-color"
	OutputFlagName                = "--output"
	ProviderFlagName              = "--provider"
	RegistryFlagName              = "--registry"
	RegistryUserFlagName          = "--registry-user"
	SetDefaultImagePrefixFlagName = "--set-default-image-prefix"
	ShellFlagname                 = "--shell"
	SubPathFlagName               = "--sub-path"
	TextFlagName                  = "--text"
)

func AllNamespacesFlag(cmd *cobra.Command, c *Config, namespace *string, allNamespaces *bool) {
	prior := cmd.PreRunE
	cmd.PreRunE = func(cmd *cobra.Command, args []string) error {
		if *allNamespaces == true {
			if cmd.Flag(StripDash(NamespaceFlagName)).Changed {
				// forbid --namespace alongside --all-namespaces
				// Check here since we need the Flag to know if the namespace came from a flag
				return ErrMultipleOneOf(NamespaceFlagName, AllNamespacesFlagName)
			}
			*namespace = ""
		}
		if prior != nil {
			if err := prior(cmd, args); err != nil {
				return err
			}
		}
		return nil
	}

	NamespaceFlag(cmd, c, namespace)
	cmd.Flags().BoolVar(allNamespaces, StripDash(AllNamespacesFlagName), false, "<todo>")
}

func NamespaceFlag(cmd *cobra.Command, c *Config, namespace *string) {
	prior := cmd.PreRunE
	cmd.PreRunE = func(cmd *cobra.Command, args []string) error {
		if *namespace == "" {
			*namespace = c.DefaultNamespace()
		}
		if prior != nil {
			if err := prior(cmd, args); err != nil {
				return err
			}
		}
		return nil
	}

	cmd.Flags().StringVarP(namespace, StripDash(NamespaceFlagName), "n", "", "<todo>")
}

func StripDash(flagName string) string {
	return strings.Replace(flagName, "--", "", 1)
}
