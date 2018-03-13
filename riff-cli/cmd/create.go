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
	"github.com/projectriff/riff/riff-cli/cmd/utils"
	"fmt"
	"errors"
)

func Create(initCmd *cobra.Command, buildCmd *cobra.Command, applyCmd *cobra.Command) (*cobra.Command, *options.CreateOptions) {
	var createOptions = options.CreateOptions{}

	createChainCmd := utils.CommandChain(initCmd, buildCmd, applyCmd)

	var createCmd = &cobra.Command{
		Use:   "create [language]",
		Short: "Create a function",
		Long:  utils.CreateCmdLong(),

		RunE:   createChainCmd.RunE,
		PreRun: createChainCmd.PreRun,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {

			//var flagset pflag.FlagSet
			//if cmd.Parent() == cmd.Root() {
			//	flagset = *cmd.PersistentFlags()
			//} else {
			//	flagset = *cmd.Parent().PersistentFlags()
			//}

			//utils.MergecreateOptions(flagset, &opts.CreateOptions.createOptions)
			//utils.MergeBuildOptions(flagset, &opts.CreateOptions)
			//utils.MergeApplyOptions(flagset, &opts.CreateOptions)

			if len(args) > 0 {
				if len(args) == 1 && createOptions.FilePath == "" {
					createOptions.FilePath = args[0]
				} else {
					return errors.New(fmt.Sprintf("Invalid argument(s) %v\n", args))
				}
			}

			err := validateInitOptions(&createOptions.InitOptions)
			if err != nil {
				return err
			}
			return nil
		},
	}

	createCmd.PersistentFlags().BoolVar(&createOptions.DryRun, "dry-run", utils.DefaultValues.DryRun, "print generated function artifacts content to stdout only")
	createCmd.PersistentFlags().StringVarP(&createOptions.FilePath, "filepath", "f", "", "path or directory used for the function resources (defaults to the current directory)")
	createCmd.PersistentFlags().StringVarP(&createOptions.FunctionName, "name", "n", "", "the name of the function (defaults to the name of the current directory)")
	createCmd.PersistentFlags().StringVar(&createOptions.RiffVersion, "riff-version", utils.DefaultValues.RiffVersion, "the version of riff to use when building containers")
	createCmd.PersistentFlags().StringVarP(&createOptions.Version, "version", "v", utils.DefaultValues.Version, "the version of the function image")
	createCmd.PersistentFlags().StringVarP(&createOptions.UserAccount, "useraccount", "u", utils.DefaultValues.UserAccount, "the Docker user account to be used for the image repository")
	createCmd.PersistentFlags().StringVarP(&createOptions.Artifact, "artifact", "a", "", "path to the function artifact, source code or jar file")
	createCmd.PersistentFlags().StringVarP(&createOptions.Input, "input", "i", "", "the name of the input topic (DefaultValues to function name)")
	createCmd.PersistentFlags().StringVarP(&createOptions.Output, "output", "o", "", "the name of the output topic (optional)")
	createCmd.PersistentFlags().BoolVar(&createOptions.Force, "force", utils.DefaultValues.Force, "overwrite existing functions artifacts")
	createCmd.PersistentFlags().BoolVarP(&createOptions.Push, "push", "", utils.DefaultValues.Push, "push the image to Docker registry")
	createCmd.PersistentFlags().StringVar(&createOptions.Namespace, "namespace", "", "the namespace used for the deployed resources (defaults to kubectl's default)")

	return createCmd, &createOptions
}

func CreateJava(initJavaCmd *cobra.Command, buildCmd *cobra.Command, applyCmd *cobra.Command, createOptions *options.CreateOptions) (*cobra.Command, *options.CreateOptions) {
	createJavaChainCmd := utils.CommandChain(initJavaCmd, buildCmd, applyCmd)
	var createJavaCmd = &cobra.Command{
		Use:   "java",
		Short: "Create a Java function",
		Long:  utils.CreateJavaCmdLong(),

		RunE: createJavaChainCmd.RunE,
		PreRun: func(cmd *cobra.Command, args []string) {
			//createOptions.Handler,_ = cmd.Flags().GetString("handler");
			createJavaChainCmd.PreRun(cmd, args)
		},
	}
	createJavaCmd.Flags().String("handler", "", "the fully qualified class name of the function handler")
	createJavaCmd.MarkFlagRequired("handler")
	return createJavaCmd, createOptions
}

func CreateShell(initShellCmd *cobra.Command, buildCmd *cobra.Command, applyCmd *cobra.Command, createOptions *options.CreateOptions) (*cobra.Command, *options.CreateOptions) {
	createShellChainCmd := utils.CommandChain(initShellCmd, buildCmd, applyCmd)
	var createShellCmd = &cobra.Command{
		Use:    "shell",
		Short:  "Create a shell script function",
		Long:   utils.CreateShellCmdLong(),
		PreRun: createShellChainCmd.PreRun,
		RunE:   createShellChainCmd.RunE,
	}
	return createShellCmd, createOptions
}

func CreateNode(initNodeCmd *cobra.Command, buildCmd *cobra.Command, applyCmd *cobra.Command, createOptions *options.CreateOptions) (*cobra.Command, *options.CreateOptions) {
	createNodeChainCmd := utils.CommandChain(initNodeCmd, buildCmd, applyCmd)
	var createNodeCmd = &cobra.Command{
		Use:     "node",
		Aliases: []string{"js"},
		Short:   "Create a node.js function",
		Long:    utils.InitNodeCmdLong(),
		PreRun:  createNodeChainCmd.PreRun,
		RunE:    createNodeChainCmd.RunE,
	}
	return createNodeCmd, createOptions
}

func CreatePython(initPythonCmd *cobra.Command, buildCmd *cobra.Command, applyCmd *cobra.Command, createOptions *options.CreateOptions) (*cobra.Command, *options.CreateOptions) {
	createPythonChainCmd := utils.CommandChain(initPythonCmd, buildCmd, applyCmd)
	var createPythonCmd = &cobra.Command{
		Use:   "python",
		Short: "Create a Python function",
		Long:  utils.InitPythonCmdLong(),

		PreRun: func(cmd *cobra.Command, args []string) {
			createOptions.Handler,_ = cmd.Flags().GetString("handler");
			if createOptions.Handler == "" {
				createOptions.Handler = createOptions.FunctionName
			}
			createPythonChainCmd.PreRun(cmd, args)
		},
		RunE: createPythonChainCmd.RunE,
	}
	createPythonCmd.Flags().String("handler", "", "the name of the function handler")
	return createPythonCmd, createOptions

}

func CreateGo(initGoCmd *cobra.Command, buildCmd *cobra.Command, applyCmd *cobra.Command, createOptions *options.CreateOptions) (*cobra.Command, *options.CreateOptions) {
	createGoChainCmd := utils.CommandChain(initGoCmd, buildCmd, applyCmd)
	var createGoCmd = &cobra.Command{
		Use:   "go",
		Short: "Create a Go function",
		Long:  utils.InitGoCmdLong(),

		PreRun: func(cmd *cobra.Command, args []string) {
			createOptions.Handler,_ = cmd.Flags().GetString("handler");
			if createOptions.Handler == "" {
				createOptions.Handler = createOptions.FunctionName
			}
			createGoChainCmd.PreRun(cmd, args)
		},
		RunE: createGoChainCmd.RunE,
	}
	createGoCmd.Flags().String("handler", "", "the name of the function handler")
	return createGoCmd, createOptions

}
