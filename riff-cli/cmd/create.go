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
)

const (
	createResult     = `create the required Dockerfile and resource definitions, and apply the resources, using sensible defaults`
	createDefinition = `Create`
)

var createOptions CreateOptions

var createChainCmd = CommandChain(initCmd, buildCmd)

var createJavaChainCmd = CommandChain(initJavaCmd, buildCmd)

var createNodeChainCmd = CommandChain(initNodeCmd, buildCmd)

var createPythonChainCmd = CommandChain(initPythonCmd, buildCmd)



var createShellChainCmd = CommandChain(initShellCmd, buildCmd)


var createCmd = &cobra.Command{
	Use:   "create [language]",
	Short: "Create a function",
	Long:  createCmdLong(initCommandDescription, LongVals{Process: createDefinition, Command: "create", Result: createResult}),

	Run: createChainCmd.Run,
	PreRun: createChainCmd.PreRun,
	PersistentPreRun: createChainCmd.PersistentPreRun,
}

var createJavaCmd = &cobra.Command{
	Use:              "java",
	Short:            "Create a Java function",
	Long:             createCmdLong(initJavaDescription, LongVals{Process: createDefinition, Command: "create java", Result: createResult}),
	Run:              createJavaChainCmd.Run,
	PreRun:           createJavaChainCmd.PreRun,
	PersistentPreRun: createJavaChainCmd.PersistentPreRun,
}

var createShellCmd = &cobra.Command{
	Use:   "shell",
	Short: "Create a shell script function",
	Long:  createCmdLong(initShellDescription, LongVals{Process: createDefinition, Command: "create shell", Result: createResult}),

	Run:              createShellChainCmd.Run,
	PreRun:           createShellChainCmd.PreRun,
	PersistentPreRun: createShellChainCmd.PersistentPreRun,
}

var createNodeCmd = &cobra.Command{
	Use:   "node",
	Short: "Create a node.js function",
	Long:  createCmdLong(initNodeDescription, LongVals{Process: createDefinition, Command: "create node", Result: createResult}),

	Run:              createNodeChainCmd.Run,
	PreRun:           createNodeChainCmd.PreRun,
	PersistentPreRun: createNodeChainCmd.PersistentPreRun,
}

var createJsCmd = &cobra.Command{
	Use:   "js",
	Short: createNodeCmd.Short,
	Long:  createNodeCmd.Long,
	Run:              createNodeChainCmd.Run,
	PreRun:           createNodeChainCmd.PreRun,
	PersistentPreRun: createNodeChainCmd.PersistentPreRun,
}


var createPythonCmd = &cobra.Command{
	Use:   "python",
	Short: "Create a Python function",
	Long:  createCmdLong(initPythonDescription, LongVals{Process: createDefinition, Command: "create python", Result: createResult}),


	Run:              createPythonChainCmd.Run,
	PreRun:           createPythonChainCmd.PreRun,
	PersistentPreRun: createPythonChainCmd.PersistentPreRun,
}

func init() {
	rootCmd.AddCommand(createCmd)
	createInitOptionFlags(createCmd, &createOptions.InitOptions)
    createBuildOptionFlags(createCmd, &createOptions.BuildOptions)

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
