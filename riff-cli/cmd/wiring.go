/*
 * Copyright 2017 the original author or authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package cmd

import (
	"github.com/spf13/cobra"
	"github.com/projectriff/riff/riff-cli/cmd/utils"
)

// CreateAndWireRootCommand creates all riff commands and sub commands, as well as the top-level 'root' command,
// wires them together and returns the root command, ready to execute.
func CreateAndWireRootCommand() *cobra.Command {

	rootCmd := Root()

	applyCmd := Apply()
	rootCmd.AddCommand(applyCmd)

	buildCmd := Build()
	rootCmd.AddCommand(buildCmd)

	rootCmd.AddCommand(List())


	initCmd := Init()
	rootCmd.AddCommand(initCmd)

	initJavaCmd := InitJava()
	initCmd.AddCommand(initJavaCmd)

	initNodeCmd := InitNode()
	initCmd.AddCommand(initNodeCmd)

	initPythonCmd := InitPython()
	initCmd.AddCommand(initPythonCmd)

	initShellCmd := InitShell()
	initCmd.AddCommand(initShellCmd)


	createChainCmd := utils.CommandChain(initCmd, buildCmd, applyCmd)
	createCmd := Create(createChainCmd)
	rootCmd.AddCommand(createCmd)

	createNodeChainCmd := utils.CommandChain(initNodeCmd, buildCmd, applyCmd)
	createCmd.AddCommand(CreateNode(createNodeChainCmd))

	createJavaChainCmd := utils.CommandChain(initJavaCmd, buildCmd, applyCmd)
	createCmd.AddCommand(CreateJava(createJavaChainCmd))

	createPythonChainCmd := utils.CommandChain(initPythonCmd, buildCmd, applyCmd)
	createCmd.AddCommand(CreatePython(createPythonChainCmd))

	createShellChainCmd := utils.CommandChain(initShellCmd, buildCmd, applyCmd)
	createCmd.AddCommand(CreateShell(createShellChainCmd))

	deleteCmd := Delete()
	rootCmd.AddCommand(deleteCmd)

	rootCmd.AddCommand(Completion(rootCmd))
	rootCmd.AddCommand(Docs(rootCmd))

	rootCmd.AddCommand(Publish())

	updateChainCmd := utils.CommandChain(buildCmd, applyCmd)
	rootCmd.AddCommand(Update(updateChainCmd))
	
	logsCmd := Logs()
	rootCmd.AddCommand(logsCmd)

	rootCmd.AddCommand(Version())
	return rootCmd
}