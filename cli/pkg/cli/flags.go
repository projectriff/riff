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
	BootstrapServersFlagName      = "--bootstrap-servers"
	CacheSizeFlagName             = "--cache-size"
	ContainerConcurrencyFlagName  = "--container-concurrency"
	ConfigFlagName                = "--config"
	ConfigurationRefFlagName      = "--configuration-ref"
	ContainerRefFlagName          = "--container-ref"
	ContentTypeFlagName           = "--content-type"
	DefaultImagePrefixFlagName    = "--default-image-prefix"
	DirectoryFlagName             = "--directory"
	DockerHubFlagName             = "--docker-hub"
	DryRunFlagName                = "--dry-run"
	EnvFlagName                   = "--env"
	EnvFromFlagName               = "--env-from"
	FunctionRefFlagName           = "--function-ref"
	GatewayFlagName               = "--gateway"
	GcrFlagName                   = "--gcr"
	GitRepoFlagName               = "--git-repo"
	GitRevisionFlagName           = "--git-revision"
	HandlerFlagName               = "--handler"
	ImageFlagName                 = "--image"
	IngressPolicyFlagName         = "--ingress-policy"
	InputFlagName                 = "--input"
	InvokerFlagName               = "--invoker"
	KubeConfigFlagName            = "--kubeconfig"
	KubeConfigFlagNameDeprecated  = "--kube-config"
	LimitCPUFlagName              = "--limit-cpu"
	LimitMemoryFlagName           = "--limit-memory"
	LocalPathFlagName             = "--local-path"
	MaxScaleFlagName              = "--max-scale"
	MinScaleFlagName              = "--min-scale"
	NamespaceFlagName             = "--namespace"
	NoColorFlagName               = "--no-color"
	OutputFlagName                = "--output"
	RegistryFlagName              = "--registry"
	RegistryUserFlagName          = "--registry-user"
	ServiceRefFlagName            = "--service-ref"
	ServiceURLFlagName            = "--service-url"
	SetDefaultImagePrefixFlagName = "--set-default-image-prefix"
	ShellFlagName                 = "--shell"
	SinceFlagName                 = "--since"
	SubPathFlagName               = "--sub-path"
	TailFlagName                  = "--tail"
	TargetPortFlagName            = "--target-port"
	WaitTimeoutFlagName           = "--wait-timeout"
)

func AllNamespacesFlag(cmd *cobra.Command, c *Config, namespace *string, allNamespaces *bool) {
	prior := cmd.PreRunE
	cmd.PreRunE = func(cmd *cobra.Command, args []string) error {
		if *allNamespaces == true {
			if cmd.Flag(StripDash(NamespaceFlagName)).Changed {
				// forbid --namespace alongside --all-namespaces
				// Check here since we need the Flag to know if the namespace came from a flag
				return ErrMultipleOneOf(NamespaceFlagName, AllNamespacesFlagName).ToAggregate()
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
	cmd.Flags().BoolVar(allNamespaces, StripDash(AllNamespacesFlagName), false, "use all kubernetes namespaces")
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

	cmd.Flags().StringVarP(namespace, StripDash(NamespaceFlagName), "n", "", "kubernetes `name`space (defaulted from kube config)")
	_ = cmd.MarkFlagCustom(StripDash(NamespaceFlagName), "__"+c.Name+"_list_namespaces")
}

func StripDash(flagName string) string {
	return strings.Replace(flagName, "--", "", 1)
}
