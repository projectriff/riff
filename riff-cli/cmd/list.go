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
	"github.com/projectriff/riff-cli/pkg/kubectl"
	"github.com/projectriff/riff-cli/pkg/ioutils"
	"github.com/spf13/viper"
)

type ListOptions struct {
	namespace string
}

var listOptions ListOptions

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List function resources",
	Long: `List the currently defined function resources.`,
	Run: func(cmd *cobra.Command, args []string) {

		// get the viper value from env var, config file or flag option
		listOptions.namespace = viper.GetString("namespace")

		fmt.Printf("Listing function resources in namespace %v\n\n", listOptions.namespace)

		cmdArgs := []string{"get", "--namespace", listOptions.namespace, "functions"}

		output, err := kubectl.ExecForString(cmdArgs)

		if err != nil {
			ioutils.Errorf("Error: %v\n", err)
			return
		}

		fmt.Printf("%v\n", output)

	},
}

func init() {
	rootCmd.AddCommand(listCmd)

	listCmd.Flags().StringP("namespace", "", "default", "the namespace used for the deployed resources")
	viper.BindPFlag("namespace", listCmd.Flags().Lookup("namespace"))
}
