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
)

var createOptions options.CreateOptions

const (
	createResult     = `create the required Dockerfile and resource definitions, and apply the resources, using sensible defaults`
	createDefinition = `Create`
)

var createChainCmd = commandChain(initCmd, buildCmd, applyCmd)

var createJavaChainCmd = commandChain(initJavaCmd, buildCmd, applyCmd)

var createNodeChainCmd = commandChain(initNodeCmd, buildCmd, applyCmd)

var createPythonChainCmd = commandChain(initPythonCmd, buildCmd, applyCmd)

var createShellChainCmd = commandChain(initShellCmd, buildCmd, applyCmd)

var createCmd = &cobra.Command{
	Use:   "create [language]",
	Short: "Create a function",
	Long:  createCmdLong(initCommandDescription, LongVals{Process: createDefinition, Command: "create", Result: createResult}),
	RunE:   createChainCmd.RunE,
	PreRun: createChainCmd.PreRun,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if !createOptions.Initialized {
			createOptions = options.CreateOptions{}
			var flagset pflag.FlagSet
			if cmd.Parent() == rootCmd {
				flagset = *cmd.PersistentFlags()
			} else {
				flagset = *cmd.Parent().PersistentFlags()
			}

			mergeInitOptions(flagset, &createOptions.InitOptions)
			mergeBuildOptions(flagset, &createOptions)
			mergeApplyOptions(flagset, &createOptions)

			if len(args) > 0 {
				if len(args) == 1 && initOptions.FunctionPath == "" {
					createOptions.FunctionPath = args[0]
				} else {
					ioutils.Errorf("Invalid argument(s) %v\n", args)
					cmd.Usage()
					os.Exit(1)
				}
			}

			err := options.ValidateAndCleanInitOptions(&createOptions.InitOptions)
			if err != nil {
				ioutils.Error(err)
				os.Exit(1)
			}
			createOptions.Initialized = true
		}
		createChainCmd.PersistentPreRun(cmd,args)
	},
}

var createJavaCmd = &cobra.Command{
	Use:              "java",
	Short:            "Create a Java function",
	Long:             createCmdLong(initJavaDescription, LongVals{Process: createDefinition, Command: "create java", Result: createResult}),
	RunE:              createJavaChainCmd.RunE,

	PreRun: func(cmd *cobra.Command, args []string) {
		handler,_ = cmd.Flags().GetString("handler")
		createJavaChainCmd.PreRun(cmd, args)
	},
	PersistentPreRun: createJavaChainCmd.PersistentPreRun,
}

var createShellCmd = &cobra.Command{
	Use:   "shell",
	Short: "Create a shell script function",
	Long:  createCmdLong(initShellDescription, LongVals{Process: createDefinition, Command: "create shell", Result: createResult}),
	PreRun: createShellChainCmd.PreRun,
	RunE:    createShellChainCmd.RunE,
}

var createNodeCmd = &cobra.Command{
	Use:   "node",
	Short: "Create a node.js function",
	Long:  createCmdLong(initNodeDescription, LongVals{Process: createDefinition, Command: "create node", Result: createResult}),
	PreRun: createNodeChainCmd.PreRun,
	RunE:    createNodeChainCmd.RunE,
}

var createJsCmd = &cobra.Command{
	Use:              "js",
	Short:            createNodeCmd.Short,
	Long:             createNodeCmd.Long,
	RunE:              createNodeChainCmd.RunE,
}

var createPythonCmd = &cobra.Command{
	Use:   "python",
	Short: "Create a Python function",
	Long:  createCmdLong(initPythonDescription, LongVals{Process: createDefinition, Command: "create python", Result: createResult}),

	PreRun: func(cmd *cobra.Command, args []string) {
		handler,_ = cmd.Flags().GetString("handler")
		createPythonChainCmd.PreRun(cmd, args)
	},
	RunE: createPythonChainCmd.RunE,
}

func init() {
	rootCmd.AddCommand(createCmd)

	createInitFlags(createCmd.PersistentFlags())
	createBuildFlags(createCmd.PersistentFlags())
	createApplyFlags(createCmd.PersistentFlags())

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