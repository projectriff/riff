/*
 * Copyright 2018 the original author or authors.
 *
 *   Licensed under the Apache License, Version 2.0 (the "License");
 *   you may not use this file except in compliance with the License.
 *   You may obtain a copy of the License at
 *
 *        https://www.apache.org/licenses/LICENSE-2.0
 *
 *   Unless required by applicable law or agreed to in writing, software
 *   distributed under the License is distributed on an "AS IS" BASIS,
 *   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *   See the License for the specific language governing permissions and
 *   limitations under the License.
 */

package cmd

import (
	"fmt"

	"github.com/projectriff/riff/riff-cli/cmd/utils"
	invoker "github.com/projectriff/riff/riff-cli/pkg/invoker"
	"github.com/projectriff/riff/riff-cli/pkg/kubectl"
	"github.com/spf13/cobra"
)

func Invokers() *cobra.Command {

	var invokersCmd = &cobra.Command{
		Use:   "invokers",
		Short: "Manage invokers in the cluster",
	}

	return invokersCmd
}

func InvokersApply(kubeCtl kubectl.KubeCtl) (*cobra.Command, *invoker.ApplyOptions) {

	invokerOperations := invoker.Operations(kubeCtl)
	var invokersApplyOptions = invoker.ApplyOptions{}

	var invokersApplyCmd = &cobra.Command{
		Use:   "apply",
		Short: "Install or update an invoker in the cluster",
		Args:  utils.AliasFlagToSoleArg("filename"),
		RunE: func(cmd *cobra.Command, args []string) error {
			out, err := invokerOperations.Apply(invokersApplyOptions)
			fmt.Print(out)
			return err
		},
	}

	invokersApplyCmd.Flags().StringVarP(&invokersApplyOptions.Filename, "filename", "f", ".", "path to the invoker resource to install")
	invokersApplyCmd.Flags().StringVarP(&invokersApplyOptions.Name, "name", "n", "", "name of the invoker (defaults to the name in the invoker resource)")
	invokersApplyCmd.Flags().StringVarP(&invokersApplyOptions.Version, "version", "v", "", "version of the invoker (defaults to the version in the invoker resource)")

	return invokersApplyCmd, &invokersApplyOptions
}

func InvokersList(kubeCtl kubectl.KubeCtl) *cobra.Command {

	invokerOperations := invoker.Operations(kubeCtl)

	var invokersListCmd = &cobra.Command{
		Use:   "list",
		Short: "List invokers in the cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			listing, err := invokerOperations.Table(args...)
			if err != nil {
				return err
			}
			fmt.Print(listing)
			return nil
		},
	}

	return invokersListCmd
}

type InvokersDeleteOptions struct {
	All  bool
	Name string
}

func InvokersDelete(kubeCtl kubectl.KubeCtl) (*cobra.Command, *InvokersDeleteOptions) {

	invokerOperations := invoker.Operations(kubeCtl)
	var invokersDeleteOptions = InvokersDeleteOptions{}

	var invokersDeleteCmd = &cobra.Command{
		Use:   "delete",
		Short: "Remove an invoker from the cluster",
		Args:  utils.AliasFlagToSoleArg("name"),
		RunE: func(cmd *cobra.Command, args []string) error {
			var out string
			var err error
			if invokersDeleteOptions.All {
				out, err = invokerOperations.DeleteAll()
			} else if invokersDeleteOptions.Name == "" {
				return fmt.Errorf("Invoker to delete must be specified")
			} else {
				out, err = invokerOperations.Delete(invokersDeleteOptions.Name)
			}
			if err != nil {
				return err
			}
			fmt.Print(out)
			return nil
		},
	}

	invokersDeleteCmd.Flags().BoolVar(&invokersDeleteOptions.All, "all", false, "remove all invokers from the cluster")
	invokersDeleteCmd.Flags().StringVarP(&invokersDeleteOptions.Name, "name", "n", "", "invoker name to remove from the cluster")

	return invokersDeleteCmd, &invokersDeleteOptions
}
