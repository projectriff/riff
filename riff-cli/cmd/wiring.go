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
)

// CreateAndWireRootCommand creates all riff commands and sub commands, as well as the top-level 'root' command,
// wires them together and returns the root command, ready to execute.
func CreateAndWireRootCommand() *cobra.Command {

	rootCmd := Root()

	initCmd, initOptions := Init()
	initJavaCmd, _ := InitJava(initOptions)
	initNodeCmd, _ := InitNode(initOptions)
	initPythonCmd, _ := InitPython(initOptions)
	initShellCmd, _ := InitShell(initOptions)
	initGoCmd, _ := InitGo(initOptions)

	initCmd.AddCommand(
		initJavaCmd,
		initGoCmd,
		initShellCmd,
		initPythonCmd,
		initNodeCmd,
	)

	buildCmd, _ := Build()

	applyCmd, _ := Apply()

	createCmd, createOptions := Create(initCmd, buildCmd, applyCmd)

	createNodeCmd, _	:= CreateNode(initNodeCmd, buildCmd, applyCmd, createOptions)
	createJavaCmd, _ 	:= CreateJava(initJavaCmd, buildCmd, applyCmd, createOptions)
	createPythonCmd, _	:= CreatePython(initPythonCmd, buildCmd, applyCmd, createOptions)
	createShellCmd, _	:= CreateShell(initShellCmd, buildCmd, applyCmd, createOptions)
	createGoCmd, _		:= CreateGo(initGoCmd, buildCmd, applyCmd, createOptions)


	createCmd.AddCommand(
		createNodeCmd,
		createJavaCmd,
		createPythonCmd,
		createShellCmd,
		createGoCmd,
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
		Update(buildCmd, applyCmd),
		Version(),
	)

	rootCmd.AddCommand(
		Completion(rootCmd),
		Docs(rootCmd),
	)

	return rootCmd
}
