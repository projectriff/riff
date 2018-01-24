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
	"github.com/projectriff/riff-cli/pkg/functions"
	"github.com/projectriff/riff-cli/pkg/ioutils"
	"github.com/projectriff/riff-cli/pkg/kubectl"
	"github.com/projectriff/riff-cli/pkg/osutils"
	"path/filepath"
)

type DeleteOptions struct {
	name string
	path string
	all  bool
}

var deleteOptions DeleteOptions

// deleteCmd represents the delete command
var deleteCmd = &cobra.Command {
	Use:   "delete",
	Short: "Delete function resources",
	Long: `Delete the resource[s] for the function or path specified.`,
	Example: `  riff delete -n square
    or
  riff delete -f function/square`,

	Run: func(cmd *cobra.Command, args []string) {

		if deleteOptions.name == "" {
			var err error
			deleteOptions.name, err = functions.FunctionNameFromPath(deleteOptions.path)
			if err != nil {
				ioutils.Errorf("Error: %v\n", err)
				return
			}
		}

		abs,err := functions.AbsPath(deleteOptions.path)
		if err != nil {
			ioutils.Errorf("Error: %v\n", err)
			return
		}

		var cmdArgs []string

		if deleteOptions.all {
			optionPath := deleteOptions.path
			if !osutils.IsDirectory(abs) {
				abs = filepath.Dir(abs)
				optionPath = filepath.Dir(optionPath)
			}
			fmt.Printf("Deleting %v resources\n\n", optionPath)
			cmdArgs = []string{"delete", "-f", abs}
		} else {
			if osutils.IsDirectory(abs) {
				fmt.Printf("Deleting %v function\n\n", deleteOptions.name)
				cmdArgs = []string{"delete", "function", deleteOptions.name}
			} else {
				fmt.Printf("Deleting %v resource\n\n", deleteOptions.path)
				cmdArgs = []string{"delete", "-f", abs}
			}
		}

		output, err := kubectl.ExecForString(cmdArgs)
		if err != nil {
			ioutils.Errorf("Error: %v\n", err)
			return
		}
		fmt.Printf("%v\n", output)

		return

	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)

	deleteCmd.Flags().StringVarP(&deleteOptions.name, "name", "n", "", "the name of the function")
	deleteCmd.Flags().StringVarP(&deleteOptions.path, "filepath", "f", "", "path or directory for the function resources, if a file is specified then the file's directory will be used (defaults to the current directory)")
	deleteCmd.Flags().BoolVarP(&deleteOptions.all, "all", "", false, "delete all resources including topics, not just the function resource")

	setFilePathFlag(deleteCmd.Flags())
}
