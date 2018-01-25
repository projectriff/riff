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
	"github.com/spf13/pflag"
	"github.com/projectriff/riff-cli/pkg/ioutils"
	"os"
	"github.com/projectriff/riff-cli/cmd/utils"
	"github.com/projectriff/riff-cli/cmd/opts"
)

var createChainCmd = utils.CommandChain(initCmd, buildCmd, applyCmd)

var createJavaChainCmd = utils.CommandChain(initJavaCmd, buildCmd, applyCmd)

var createNodeChainCmd = utils.CommandChain(initNodeCmd, buildCmd, applyCmd)

var createPythonChainCmd = utils.CommandChain(initPythonCmd, buildCmd, applyCmd)

var createShellChainCmd = utils.CommandChain(initShellCmd, buildCmd, applyCmd)

var createCmd = &cobra.Command{
	Use:   "create [language]",
	Short: "Create a function",
	Long:  utils.CreateCmdLong(),

	RunE:   createChainCmd.RunE,
	PreRun: createChainCmd.PreRun,
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
		createChainCmd.PersistentPreRun(cmd, args)
	},
}

var createJavaCmd = &cobra.Command{
	Use:   "java",
	Short: "Create a Java function",
	Long:  utils.CreateJavaCmdLong(),

	RunE: createJavaChainCmd.RunE,
	PreRun: func(cmd *cobra.Command, args []string) {
		opts.Handler = utils.GetHandler(cmd)
		createJavaChainCmd.PreRun(cmd, args)
	},
	PersistentPreRun: createJavaChainCmd.PersistentPreRun,
}

var createShellCmd = &cobra.Command{
	Use:   "shell",
	Short: "Create a shell script function",
	Long:  utils.CreateShellCmdLong(),

	PreRun: createShellChainCmd.PreRun,
	RunE:   createShellChainCmd.RunE,
}

var createNodeCmd = &cobra.Command{
	Use:   "node",
	Short: "Create a node.js function",
	Long:  utils.InitNodeCmdLong(),

	PreRun: createNodeChainCmd.PreRun,
	RunE:   createNodeChainCmd.RunE,
}

var createJsCmd = &cobra.Command{
	Use:   "js",
	Short: createNodeCmd.Short,
	Long:  createNodeCmd.Long,
	RunE:  createNodeChainCmd.RunE,
}

var createPythonCmd = &cobra.Command{
	Use:   "python",
	Short: "Create a Python function",
	Long:  utils.InitPythonCmdLong(),

	PreRun: func(cmd *cobra.Command, args []string) {
		opts.Handler = utils.GetHandler(cmd)
		createPythonChainCmd.PreRun(cmd, args)
	},
	RunE: createPythonChainCmd.RunE,
}

func init() {
	rootCmd.AddCommand(createCmd)

	utils.CreateInitFlags(createCmd.PersistentFlags())
	utils.CreateBuildFlags(createCmd.PersistentFlags())
	utils.CreateApplyFlags(createCmd.PersistentFlags())

	createCmd.AddCommand(createJavaCmd)
	createCmd.AddCommand(createJsCmd)
	createCmd.AddCommand(createNodeCmd)
	createCmd.AddCommand(createPythonCmd)
	createCmd.AddCommand(createShellCmd)

	createJavaCmd.Flags().String("handler", "", "the fully qualified class name of the function handler")
	createJavaCmd.MarkFlagRequired("handler")

	createPythonCmd.Flags().String("handler", "", "the name of the function handler")
	createPythonCmd.MarkFlagRequired("handler")
}
