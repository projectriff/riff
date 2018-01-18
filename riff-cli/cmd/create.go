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
"github.com/projectriff/riff-cli/pkg/osutils"
"github.com/projectriff/riff-cli/pkg/ioutils"
)


const (
	createResult = `create the required Dockerfile and resource definitions, and apply the resources, using sensible defaults`
	createDefinition = `Create`
)

var createOptions CreateOptions

var createCmd = &cobra.Command{
	Use:   "create [language]",
	Short: "Create a function",
	Long: createCmdLong(initCommandDescription, LongVals{Process: createDefinition, Command:"create", Result:createResult}),

	Run: func(cmd *cobra.Command, args []string) {

	},
	PersistentPreRun: initCmd.PersistentPreRun,
}

var createJavaCmd = &cobra.Command{
	Use:   "java",
	Short: "Create a Java function",
	Long: 	createCmdLong(initJavaDescription, LongVals{Process:createDefinition, Command:"create java", Result:createResult}),
	Run: func(cmd *cobra.Command, args []string) {

	},
}

var createShellCmd = &cobra.Command{
	Use:   "shell",
	Short: "Create a shell script function",
	Long: 	createCmdLong(initShellDescription, LongVals{Process:createDefinition, Command:"create shell", Result:createResult}),

	Run: func(cmd *cobra.Command, args []string) {
		initializer := NewShellInitializer()
		err := initializer.initialize(initOptions)
		if err != nil {
			ioutils.Error(err)
			return
		}
	},
}

var createNodeCmd = &cobra.Command{
	Use:   "node",
	Short: "Create a node.js function",
	Long: 	createCmdLong(initNodeDescription, LongVals{Process:createDefinition, Command:"create node", Result:createResult}),

	Run: createNode,
}

var createJsCmd = &cobra.Command{
	Use:   "js",
	Short: initNodeCmd.Short,
	Long:  initNodeCmd.Long,
	Run:   initNodeCmd.Run,
}

func createNode(cmd *cobra.Command, args []string) {

}

var createPythonCmd = &cobra.Command{
	Use:   "python",
	Short: "Create a Python function",
	Long: createCmdLong(initPythonDescription, LongVals{Process:createDefinition, Command:"create python", Result:createResult}),


	Run: func(cmd *cobra.Command, args []string) {

	},
}


func init() {
	rootCmd.AddCommand(createCmd)
	createCmd.PersistentFlags().AddFlagSet(initCmd.PersistentFlags())
	createCmd.PersistentFlags().StringVarP(&createOptions.userAccount, "useraccount", "u", osutils.GetCurrentUsername(), "the Docker user account to be used for the image repository (defaults to current OS username")
	createCmd.PersistentFlags().StringVarP(&createOptions.riffVersion, "riff-version", "", "0.0.1", "the version of riff to use when building containers")
	createCmd.PersistentFlags().BoolVarP(&createOptions.push, "push", "", false, "push the image to Docker registry")

	createCmd.AddCommand(createJavaCmd)
	createCmd.AddCommand(createJsCmd)
	createCmd.AddCommand(createNodeCmd)
	createCmd.AddCommand(createPythonCmd)
	createCmd.AddCommand(createShellCmd)
	createJavaCmd.Flags().AddFlagSet(initCmd.Flags())
	createJavaCmd.MarkFlagRequired("handler")
	createPythonCmd.Flags().AddFlagSet(initCmd.Flags())
	createPythonCmd.MarkFlagRequired("handler")
}

