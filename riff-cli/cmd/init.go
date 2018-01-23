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
	"github.com/projectriff/riff-cli/pkg/ioutils"
	"github.com/projectriff/riff-cli/pkg/options"
	"os"
	"github.com/spf13/pflag"
	"github.com/projectriff/riff-cli/cmd/utils"
	"github.com/projectriff/riff-cli/cmd/opts"
	"github.com/projectriff/riff-cli/pkg/initializers"
)



/*
 * init Command
 * TODO: Use cmd.Example
 */


var initCmd = &cobra.Command{
	Use:   "init [language]",
	Short: "Initialize a function",
	Long:  	utils.InitCmdLong(),

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
			if cmd.Parent() == rootCmd {
				flagset = *cmd.PersistentFlags()
			} else {
				flagset = *cmd.Parent().PersistentFlags()
			}
			utils.MergeInitOptions(flagset, &opts.InitOptions)

			if len(args) > 0 {
				if len(args) == 1 && opts.InitOptions.FunctionPath == "" {
					opts.InitOptions.FunctionPath = args[0]
				} else {
					ioutils.Errorf("Invalid argument(s) %v\n", args)
					cmd.Usage()
					os.Exit(1)
				}
			}

			err := options.ValidateAndCleanInitOptions(&opts.InitOptions)
			if err != nil {
				os.Exit(1)
			}

			opts.CreateOptions.Initialized = true
		}
	},
}

/*
 * init java Command
 */


var initJavaCmd = &cobra.Command{
	Use:   "java",
	Short: "Initialize a Java function",
	Long: 	utils.InitJavaCmdLong(),
	RunE: func(cmd *cobra.Command, args []string) error {
		opts.InitOptions.Handler = utils.GetHandler(cmd)
		err := initializers.Java().Initialize(opts.InitOptions)
		if err != nil {
			return err
		}
		return nil
	},
}
/*
 * init shell ommand
 */


var initShellCmd = &cobra.Command{
	Use:   "shell",
	Short: "Initialize a shell script function",
	Long:	utils.InitShellCmdLong(),

	RunE: func(cmd *cobra.Command, args []string) error {
		err := initializers.Shell().Initialize(opts.InitOptions)
		if err != nil {
			return err
		}
		return nil
	},
}
/*
 * init node Command
 */

var initNodeCmd = &cobra.Command{
	Use:   "node",
	Short: "Initialize a node.js function",
	Long:	utils.InitNodeCmdLong(),

	RunE: func(cmd *cobra.Command, args []string) error {
		err := initializers.Node().Initialize(opts.InitOptions)
		if err != nil {
			return err
		}
		return nil
	},
	Aliases: []string{"js"},
}

/*
 * init python Command
 */


var initPythonCmd = &cobra.Command{
	Use:   "python",
	Short: "Initialize a Python function",
	Long:	utils.InitPythonCmdLong(),

	RunE: func(cmd *cobra.Command, args []string) error {
		opts.InitOptions.Handler = utils.GetHandler(cmd)
		err := initializers.Python().Initialize(opts.InitOptions)
		if err != nil {
			return err
		}
		return nil
	},
}


func init() {

	rootCmd.AddCommand(initCmd)

	utils.CreateInitFlags(initCmd.PersistentFlags())

	initCmd.AddCommand(initJavaCmd)
	initCmd.AddCommand(initNodeCmd)
	initCmd.AddCommand(initPythonCmd)
	initCmd.AddCommand(initShellCmd)

	initJavaCmd.Flags().String("handler", "", "the fully qualified class name of the function handler")
	initJavaCmd.MarkFlagRequired("handler")

	initPythonCmd.Flags().String("handler", "", "the name of the function handler")
	initPythonCmd.MarkFlagRequired("handler")
}
