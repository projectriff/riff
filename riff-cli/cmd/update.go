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
	"github.com/spf13/cobra"
	"github.com/projectriff/riff-cli/pkg/options"
	"github.com/projectriff/riff-cli/pkg/ioutils"
	"github.com/projectriff/riff-cli/cmd/utils"
	"github.com/projectriff/riff-cli/cmd/opts"
	"github.com/spf13/pflag"
	"os"
)

var updateChainCmd = utils.CommandChain(buildCmd, applyCmd)

// updateCmd represents the update command
var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update a function",
	Long: `Build the function based on the code available in the path directory, using the name and version specified 
  for the image that is built. Then Apply the resource definition[s] included in the path.`,
	Example: `  riff update -n <name> -v <version> -f <path> [--push]`,

	RunE:   updateChainCmd.RunE,
	PreRun: updateChainCmd.PreRun,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if !opts.CreateOptions.Initialized {
			opts.CreateOptions = options.CreateOptions{}
			var flagset pflag.FlagSet
			if cmd.Parent() == rootCmd {
				flagset = *cmd.PersistentFlags()
			} else {
				flagset = *cmd.Parent().PersistentFlags()
			}

			utils.MergeInitOptions(flagset, &opts.CreateOptions.InitOptions)
			utils.MergeBuildOptions(flagset, &opts.CreateOptions)
			utils.MergeApplyOptions(flagset, &opts.CreateOptions)

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
			opts.CreateOptions.Initialized = true
		}
		updateChainCmd.PersistentPreRun(cmd, args)
	},
}


func init() {
	rootCmd.AddCommand(updateCmd)
	utils.CreateBuildFlags(updateCmd.PersistentFlags())
	utils.CreateApplyFlags(updateCmd.PersistentFlags())
}
