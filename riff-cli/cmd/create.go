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
	"github.com/projectriff/riff/riff-cli/pkg/options"
	"github.com/spf13/pflag"
	"github.com/projectriff/riff/riff-cli/pkg/ioutils"
	"os"
	"github.com/projectriff/riff/riff-cli/cmd/utils"
	"github.com/projectriff/riff/riff-cli/cmd/opts"
)


func Create(createChainCmd *cobra.Command) *cobra.Command {
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
				if cmd.Parent() == cmd.Root() {
					flagset = *cmd.PersistentFlags()
				} else {
					flagset = *cmd.Parent().PersistentFlags()
				}

				utils.MergeInitOptions(flagset, &opts.CreateOptions.InitOptions)
				utils.MergeBuildOptions(flagset, &opts.CreateOptions)
				utils.MergeApplyOptions(flagset, &opts.CreateOptions)

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
				opts.CreateOptions.Initialized = true
			}
			createChainCmd.PersistentPreRun(cmd, args)
		},
	}

	utils.CreateInitFlags(createCmd.PersistentFlags())
	utils.CreateBuildFlags(createCmd.PersistentFlags())
	utils.CreateApplyFlags(createCmd.PersistentFlags())

	return createCmd
}

func CreateJava(createJavaChainCmd *cobra.Command) *cobra.Command {
	var createJavaCmd = &cobra.Command{
		Use:   "java",
		Short: "Create a Java function",
		Long:  utils.CreateJavaCmdLong(),

		RunE: createJavaChainCmd.RunE,
		PreRun: func(cmd *cobra.Command, args []string) {
			opts.Handler = utils.GetHandler(cmd)
			createJavaChainCmd.PreRun(cmd, args)
		},
	}
	createJavaCmd.Flags().String("handler", "", "the fully qualified class name of the function handler")
	createJavaCmd.MarkFlagRequired("handler")
	return createJavaCmd
}

func CreateShell(createShellChainCmd *cobra.Command) *cobra.Command {
	var createShellCmd = &cobra.Command{
		Use:    "shell",
		Short:  "Create a shell script function",
		Long:   utils.CreateShellCmdLong(),
		PreRun: createShellChainCmd.PreRun,
		RunE:   createShellChainCmd.RunE,
	}
	return createShellCmd
}

func CreateNode(createNodeChainCmd *cobra.Command) *cobra.Command {
	var createNodeCmd = &cobra.Command{
		Use:     "node",
		Aliases: []string{"js"},
		Short:   "Create a node.js function",
		Long:    utils.InitNodeCmdLong(),
		PreRun:  createNodeChainCmd.PreRun,
		RunE:    createNodeChainCmd.RunE,
	}
	return createNodeCmd
}

func CreatePython(createPythonChainCmd *cobra.Command) *cobra.Command {
	var createPythonCmd = &cobra.Command{
		Use:   "python",
		Short: "Create a Python function",
		Long:  utils.InitPythonCmdLong(),

		PreRun: func(cmd *cobra.Command, args []string) {
			opts.Handler = utils.GetHandler(cmd)
			if opts.Handler == "" {
				opts.Handler = opts.CreateOptions.FunctionName
			}
			createPythonChainCmd.PreRun(cmd, args)
		},
		RunE: createPythonChainCmd.RunE,
	}
	createPythonCmd.Flags().String("handler", "", "the name of the function handler")
	return createPythonCmd

}

func CreateGo(createGoChainCmd *cobra.Command) *cobra.Command {
	var createGoCmd = &cobra.Command{
		Use:   "go",
		Short: "Create a Go function",
		Long:  utils.InitGoCmdLong(),

		PreRun: func(cmd *cobra.Command, args []string) {
			opts.Handler = utils.GetHandler(cmd)
			if opts.Handler == "" {
				opts.Handler = opts.CreateOptions.FunctionName
			}
			createGoChainCmd.PreRun(cmd, args)
		},
		RunE: createGoChainCmd.RunE,
	}
	createGoCmd.Flags().String("handler", "", "the name of the function handler")
	return createGoCmd

}