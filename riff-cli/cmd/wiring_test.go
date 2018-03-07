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

import "github.com/spf13/cobra"

// Temporary global variables and workarounds so that tests as they were written continue to work

var (
	rootCmd = CreateAndWireRootCommand()

	initCmd         = locate("init")
	createCmd       = locate("create")
	initJavaCmd     = locate("init", "java")
	createJavaCmd   = locate("create", "java")
	initPythonCmd   = locate("init", "python")
	createPythonCmd = locate("create", "python")
	createShellCmd  = locate("create", "shell")
	createNodeCmd   = locate("create", "node")
	deleteCmd       = locate("delete")
)

func locate(a ... string) *cobra.Command {
	command, _, err := rootCmd.Find(a)
	if err != nil {
		panic(err)
	}
	return command
}
