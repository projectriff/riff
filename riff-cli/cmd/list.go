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

	listCmd.Flags().StringVarP(&listOptions.namespace, "namespace", "", "default", "the namespace used for the deployed resources")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// listCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// listCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
