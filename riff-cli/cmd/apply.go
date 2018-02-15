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
	"github.com/projectriff/riff-cli/pkg/options"
	"github.com/projectriff/riff-cli/cmd/utils"
	"github.com/projectriff/riff-cli/cmd/opts"
	"github.com/projectriff/riff-cli/pkg/functions"
	"github.com/projectriff/riff-cli/pkg/osutils"
	"strings"
)

// applyCmd represents the apply command
var applyCmd = &cobra.Command{
	Use:   "apply",
	Short: "Apply function resource definitions",
	Long: `Apply the resource definition[s] included in the path. A resource will be created if it doesn't exist yet.`,
  Example: `  riff apply -f some/function/path
  riff apply -f some/function/path/some.yaml`,

	RunE: func(cmd *cobra.Command, args []string) error {
		return apply(cmd, options.GetApplyOptions(opts.CreateOptions))
	},
	PreRun: func(cmd *cobra.Command, args []string) {

		if !opts.CreateOptions.Initialized {
			utils.MergeApplyOptions(*cmd.Flags(), &opts.CreateOptions)
			if len(args) > 0 {
				if len(args) == 1 && opts.CreateOptions.FilePath == "" {
					opts.CreateOptions.FilePath = args[0]
				} else {
					ioutils.Errorf("Invalid argument(s) %v\n", args)
					cmd.Usage()
					os.Exit(1)
				}
			}

			err := options.ValidateAndCleanInitOptions(&opts.CreateOptions.InitOptions)
			if err != nil {
				ioutils.Error(err)
				os.Exit(1)
			}
		}
		opts.CreateOptions.Initialized = true
	},
}

func apply(cmd *cobra.Command, opts options.ApplyOptions) error {
	//fnDir, _ := functions.FunctionDirFromPath(opts.FilePath)
	abs,err := functions.AbsPath(opts.FilePath)
	if err != nil {
		cmd.SilenceUsage = true
		return err
	}

	var cmdArgs []string
	var message string

	if osutils.IsDirectory(abs) {
		message = fmt.Sprintf("Applying resources in %v\n\n", opts.FilePath)
	} else {
		message = fmt.Sprintf("Applying resource %v\n\n", opts.FilePath)
	}
	cmdArgs = []string{"apply", "--namespace", opts.Namespace, "-f", abs}


	if opts.DryRun {
		fmt.Printf("\nApply Command: kubectl %s\n\n", strings.Trim(fmt.Sprint(cmdArgs), "[]"))
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
	rootCmd.AddCommand(applyCmd)
	utils.CreateApplyFlags(applyCmd.Flags())
}
