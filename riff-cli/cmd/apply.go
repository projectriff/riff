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
	"os"
)

var applyFilePath string

// applyCmd represents the apply command
var applyCmd = &cobra.Command{
	Use:   "apply",
	Short: "Apply function resource definitions",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		output, err := kubectl.ExecForString([]string{"apply", "-f", applyFilePath})
		if err != nil {
			return
		}
		fmt.Println(output)
	},
	PreRun: func(cmd *cobra.Command, args []string) {
		if len(args) > 0 {
			if len(args) == 1 && applyFilePath == "" {
				applyFilePath = args[0]
			} else {
				ioutils.Errorf("Invalid argument(s) %v\n", args)
				cmd.Usage()
				os.Exit(1)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(applyCmd)
	applyCmd.Flags().StringVarP(&applyFilePath, "filepath", "f", "", "Filename, directory, or URL to files that contains the configuration to apply")
}
