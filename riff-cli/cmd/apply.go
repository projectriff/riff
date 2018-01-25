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
)

// applyCmd represents the apply command
var applyCmd = &cobra.Command{
	Use:   "apply",
	Short: "Apply function resource definitions",
	Long: `Apply the resource definition[s] included in the path. A resource will be created if it doesn't exist yet.`,
  Example: `
riff apply -f some/function/path
riff apply -f some/function/path/some.yaml
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return apply(cmd, options.GetApplyOptions(opts.CreateOptions))
	},
	PreRun: func(cmd *cobra.Command, args []string) {

		if !opts.CreateOptions.Initialized {
			utils.MergeApplyOptions(*cmd.Flags(), &opts.CreateOptions)
			if len(args) > 0 {
				if len(args) == 1 && opts.CreateOptions.FunctionPath == "" {
					opts.CreateOptions.FunctionPath = args[0]
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
	if opts.DryRun {
		fmt.Printf("\nApply Command: kubectl apply -f %s\n\n", opts.FunctionPath)
	} else {
		output, err := kubectl.ExecForString([]string{"apply", "-f", opts.FunctionPath})
		if err != nil {
			cmd.SilenceUsage = true
			return err
		}
		fmt.Println(output)
	}
	return nil
}

func init() {
	rootCmd.AddCommand(applyCmd)
	utils.CreateApplyFlags(applyCmd.Flags())
}
