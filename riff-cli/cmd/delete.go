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

	"os"
	"path/filepath"
	"strings"

	"github.com/projectriff/riff-cli/cmd/utils"
	"github.com/projectriff/riff-cli/pkg/functions"
	"github.com/projectriff/riff-cli/pkg/ioutils"
	"github.com/projectriff/riff-cli/pkg/kubectl"
	"github.com/projectriff/riff-cli/pkg/options"
	"github.com/projectriff/riff-cli/pkg/osutils"
	"github.com/spf13/cobra"
)

var DeleteAllOptions options.DeleteAllOptions

// deleteCmd represents the delete command
var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete function resources",
	Long:  `Delete the resource[s] for the function or path specified.`,
	Example: `  riff delete -n square
    or
  riff delete -f function/square`,

	RunE: func(cmd *cobra.Command, args []string) error {

		return delete(cmd, options.GetDeleteOptions(DeleteAllOptions))

	},
	PreRun: func(cmd *cobra.Command, args []string) {

		if !DeleteAllOptions.Initialized {
			utils.MergeDeleteOptions(*cmd.Flags(), &DeleteAllOptions)
			if len(args) > 0 {
				if len(args) == 1 && DeleteAllOptions.FilePath == "" {
					DeleteAllOptions.FilePath = args[0]
				} else {
					ioutils.Errorf("Invalid argument(s) %v\n", args)
					cmd.Usage()
					os.Exit(1)
				}
			}

			err := options.ValidateNamePathOptions(&DeleteAllOptions.FunctionName, &DeleteAllOptions.FilePath)
			if err != nil {
				ioutils.Error(err)
				os.Exit(1)
			}
		}
		DeleteAllOptions.Initialized = true
	},
}

func delete(cmd *cobra.Command, opts options.DeleteOptions) error {

	if opts.FunctionName == "" {
		var err error
		opts.FunctionName, err = functions.FunctionNameFromPath(opts.FilePath)
		if err != nil {
			cmd.SilenceUsage = true
			return err
		}
	}

	abs, err := functions.AbsPath(opts.FilePath)
	if err != nil {
		cmd.SilenceUsage = true
		return err
	}

	var cmdArgs []string
	var message string

	if opts.All {
		optionPath := opts.FilePath
		if !osutils.IsDirectory(abs) {
			abs = filepath.Dir(abs)
			optionPath = filepath.Dir(optionPath)
		}
		message = fmt.Sprintf("Deleting %v resources\n\n", optionPath)
		resourceDefinitionPaths, err := osutils.FindRiffResourceDefinitionPaths(abs)
		if err != nil {
			return err
		}
		cmdArgs = []string{"delete", "--namespace", opts.Namespace}
		for _, resourceDefinitionPath := range resourceDefinitionPaths {
			cmdArgs = append(cmdArgs, "-f", resourceDefinitionPath)
		}
	} else {
		if osutils.IsDirectory(abs) {
			message = fmt.Sprintf("Deleting %v function\n\n", opts.FunctionName)
			cmdArgs = []string{"delete", "--namespace", opts.Namespace, "function", opts.FunctionName}
		} else {
			message = fmt.Sprintf("Deleting %v resource\n\n", opts.FilePath)
			cmdArgs = []string{"delete", "--namespace", opts.Namespace, "-f", opts.FilePath}
		}
	}

	if opts.DryRun {
		fmt.Printf("\nDelete Command: kubectl %s\n\n", strings.Trim(fmt.Sprint(cmdArgs), "[]"))
	} else {
		fmt.Print(message)
		output, err := kubectl.ExecForString(cmdArgs)
		if err != nil {
			cmd.SilenceUsage = true
			return err
		}
		fmt.Printf("%v\n", output)
	}

	return nil
}

func init() {
	rootCmd.AddCommand(deleteCmd)
	utils.CreateDeleteFlags(deleteCmd.Flags())
}
