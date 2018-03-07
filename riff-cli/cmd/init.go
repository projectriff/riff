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
	"github.com/projectriff/riff/riff-cli/pkg/ioutils"
	"github.com/projectriff/riff/riff-cli/pkg/options"
	"os"
	"github.com/spf13/pflag"
	"github.com/projectriff/riff/riff-cli/cmd/utils"
	"github.com/projectriff/riff/riff-cli/cmd/opts"
	"github.com/projectriff/riff/riff-cli/pkg/initializers"
)

func Init() *cobra.Command {

	var initCmd = &cobra.Command{
		Use:   "init [language]",
		Short: "Initialize a function",
		Long:  utils.InitCmdLong(),

		RunE: func(cmd *cobra.Command, args []string) error {
			err := initializers.Initialize(opts.InitOptions)
			if err != nil {
				return err
			}
			return nil
		},
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			opts.InitOptions = opts.CreateOptions.InitOptions
			if !opts.CreateOptions.Initialized {
				opts.InitOptions = opts.CreateOptions.InitOptions
				var flagset pflag.FlagSet
				if cmd.Parent() == cmd.Root()  {
					flagset = *cmd.PersistentFlags()
				} else {
					flagset = *cmd.Parent().PersistentFlags()
				}
				utils.MergeInitOptions(flagset, &opts.InitOptions)

				if len(args) > 0 {
					if len(args) == 1 && opts.InitOptions.FilePath == "" {
						opts.InitOptions.FilePath = args[0]
					} else {
						ioutils.Errorf("Invalid argument(s) %v\n", args)
						cmd.Usage()
						os.Exit(1)
					}
				}

				err := options.ValidateAndCleanInitOptions(&opts.InitOptions)
				if err != nil {
					ioutils.Error(err)
					os.Exit(1)
				}

				opts.CreateOptions.Initialized = true
			}
		},
	}
	utils.CreateInitFlags(initCmd.PersistentFlags())

	return initCmd
}

func InitJava() *cobra.Command {
	var initJavaCmd = &cobra.Command{
		Use:   "java",
		Short: "Initialize a Java function",
		Long:  utils.InitJavaCmdLong(),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.InitOptions.Handler = utils.GetHandler(cmd)
			err := initializers.Java().Initialize(opts.InitOptions)
			if err != nil {
				return err
			}
			return nil
		},
	}
	initJavaCmd.Flags().String("handler", "", "the fully qualified class name of the function handler")
	initJavaCmd.MarkFlagRequired("handler")

	return initJavaCmd
}

func InitShell() *cobra.Command {
	var initShellCmd = &cobra.Command{
		Use:   "shell",
		Short: "Initialize a shell script function",
		Long:  utils.InitShellCmdLong(),

		RunE: func(cmd *cobra.Command, args []string) error {
			err := initializers.Shell().Initialize(opts.InitOptions)
			if err != nil {
				return err
			}
			return nil
		},
	}
	return initShellCmd
}

func InitNode() *cobra.Command {
	var initNodeCmd = &cobra.Command{
		Use:   "node",
		Short: "Initialize a node.js function",
		Long:  utils.InitNodeCmdLong(),

		RunE: func(cmd *cobra.Command, args []string) error {
			err := initializers.Node().Initialize(opts.InitOptions)
			if err != nil {
				return err
			}
			return nil
		},
		Aliases: []string{"js"},
	}
	return initNodeCmd
}

func InitPython() *cobra.Command {
	var initPythonCmd = &cobra.Command{
		Use:   "python",
		Short: "Initialize a Python function",
		Long:  utils.InitPythonCmdLong(),

		RunE: func(cmd *cobra.Command, args []string) error {
			opts.InitOptions.Handler = utils.GetHandler(cmd)
			if opts.InitOptions.Handler == "" {
				opts.InitOptions.Handler = opts.InitOptions.FunctionName
			}
			err := initializers.Python().Initialize(opts.InitOptions)
			if err != nil {
				return err
			}
			return nil
		},
	}
	initPythonCmd.Flags().String("handler", "", "the name of the function handler")
	return initPythonCmd
}
