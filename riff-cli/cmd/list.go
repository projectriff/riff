/*
 * Copyright 2018 the original author or authors.
 *
 *   Licensed under the Apache License, Version 2.0 (the "License");
 *   you may not use this file except in compliance with the License.
 *   You may obtain a copy of the License at
 *  
 *        http://www.apache.org/licenses/LICENSE-2.0
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

	"github.com/spf13/cobra"
	"github.com/projectriff/riff/riff-cli/pkg/kubectl"
	"github.com/projectriff/riff/riff-cli/cmd/utils"
)

type ListOptions struct {
	namespace string
}

func List() *cobra.Command {

	var listOptions ListOptions
	// listCmd represents the list command
	var listCmd = &cobra.Command{
		Use:   "list",
		Short: "List function resources",
		Long:  `List the currently defined function resources.`,
		RunE: func(cmd *cobra.Command, args []string) error {

			// get the viper value from env var, config file or flag option
			listOptions.namespace = utils.GetStringValueWithOverride("namespace", *cmd.Flags())

			if listOptions.namespace != "" {
				fmt.Printf("Listing function resources in namespace %v\n\n", listOptions.namespace)
			} else {
				fmt.Print("Listing function resources\n\n")

			}

			cmdArgs := []string{"get"}
			if listOptions.namespace != "" {
				cmdArgs = append(cmdArgs, "--namespace", listOptions.namespace)
			}
			cmdArgs = append(cmdArgs, "functions")

			output, err := kubectl.ExecForString(cmdArgs)

			if err != nil {
				return err
			}

			fmt.Printf("%v\n", output)
			return nil

		},
	}

	listCmd.Flags().StringP("namespace", "", "", "the namespace used for the deployed resources")
	return listCmd
}
