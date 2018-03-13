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
	"github.com/projectriff/riff/riff-cli/cmd/utils"
)

func Create(initCmd *cobra.Command, buildCmd *cobra.Command, applyCmd *cobra.Command) (*cobra.Command) {

	createChainCmd := utils.CommandChain(initCmd, buildCmd, applyCmd)
	createChainCmd.Use =  "create [language]"
	createChainCmd.Short = "Create a function"
	createChainCmd.Long =  utils.CreateCmdLong()
	return createChainCmd
}

func CreateJava(initJavaCmd *cobra.Command, buildCmd *cobra.Command, applyCmd *cobra.Command) (*cobra.Command) {
	createJavaChainCmd := utils.CommandChain(initJavaCmd, buildCmd, applyCmd)
	createJavaChainCmd.Use = "java"
	createJavaChainCmd.Short = "Create a Java function"
	createJavaChainCmd.Long =  utils.CreateJavaCmdLong()
	return createJavaChainCmd
}

func CreateShell(initShellCmd *cobra.Command, buildCmd *cobra.Command, applyCmd *cobra.Command) (*cobra.Command) {
	createShellChainCmd := utils.CommandChain(initShellCmd, buildCmd, applyCmd)
	createShellChainCmd.Use = "shell"
	createShellChainCmd.Short = "Create a shell script function"
	createShellChainCmd.Long =  utils.CreateShellCmdLong()

	return createShellChainCmd
}

func CreateNode(initNodeCmd *cobra.Command, buildCmd *cobra.Command, applyCmd *cobra.Command) (*cobra.Command) {
	createNodeChainCmd := utils.CommandChain(initNodeCmd, buildCmd, applyCmd)
	createNodeChainCmd.Use = "node"
	createNodeChainCmd.Aliases = []string{"js"}
	createNodeChainCmd.Short =  "Create a node.js function"
	createNodeChainCmd.Long = utils.CreateNodeCmdLong()
	return createNodeChainCmd
}

func CreatePython(initPythonCmd *cobra.Command, buildCmd *cobra.Command, applyCmd *cobra.Command) (*cobra.Command) {
	createPythonChainCmd := utils.CommandChain(initPythonCmd, buildCmd, applyCmd)
	createPythonChainCmd.Use = "python"
	createPythonChainCmd.Short = "Create a Python function"
	createPythonChainCmd.Long =  utils.CreatePythonCmdLong()
	return createPythonChainCmd

}

func CreateGo(initGoCmd *cobra.Command, buildCmd *cobra.Command, applyCmd *cobra.Command) (*cobra.Command) {
	createGoChainCmd := utils.CommandChain(initGoCmd, buildCmd, applyCmd)
	createGoChainCmd.Use = "go"
	createGoChainCmd.Short = "Create a Go function"
	createGoChainCmd.Long = utils.InitGoCmdLong()
	return createGoChainCmd

}
