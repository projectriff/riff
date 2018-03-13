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

	initCmd := Init()
	initJavaCmd := InitJava()
	initNodeCmd := InitNode()
	initPythonCmd := InitPython()
	initShellCmd := InitShell()
	initGoCmd := InitGo()
	initCmd.AddCommand(
		initJavaCmd,
		initGoCmd,
		initShellCmd,
		initPythonCmd,
		initNodeCmd,
	)

	buildCmd := Build()

	applyCmd := Apply()

	createCmd := Create(utils.CommandChain(initCmd, buildCmd, applyCmd))
	createCmd.AddCommand(
		CreateNode(utils.CommandChain(initNodeCmd, buildCmd, applyCmd)),
		CreateJava(utils.CommandChain(initJavaCmd, buildCmd, applyCmd)),
		CreatePython(utils.CommandChain(initPythonCmd, buildCmd, applyCmd)),
		CreateShell(utils.CommandChain(initShellCmd, buildCmd, applyCmd)),
		CreateGo(utils.CommandChain(initGoCmd, buildCmd, applyCmd)),
	)

	deleteCmd, _ := Delete()

	rootCmd.AddCommand(
		applyCmd,
		buildCmd,
		createCmd,
		deleteCmd,
		initCmd,
		List(),
		Logs(),
		Publish(),
		Update(utils.CommandChain(buildCmd, applyCmd)),
		Version(),
	)


	rootCmd.AddCommand(
		Completion(rootCmd),
		Docs(rootCmd),
	)

	return rootCmd
}