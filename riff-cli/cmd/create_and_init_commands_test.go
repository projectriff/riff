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
	"testing"
	"github.com/stretchr/testify/assert"
	"github.com/projectriff/riff/riff-cli/pkg/options"

	"os"
	"github.com/projectriff/riff/riff-cli/pkg/osutils"
	"fmt"
	"github.com/spf13/cobra"
)

func TestCreateCommandImplicitPath(t *testing.T) {

	rootCmd, createOptions := setupCreateTest()

	as := assert.New(t)
	rootCmd.SetArgs([]string{"create", "--dry-run", "../test_data/shell/echo", "-v", "0.0.1-snapshot"})

	_, err := rootCmd.ExecuteC()
	as.NoError(err)

	as.NotEmpty(createOptions.FilePath)
	as.Equal("../test_data/shell/echo", createOptions.FilePath)

}

func TestCreateCommandFromCWD(t *testing.T) {

	rootCmd, _ := setupCreateTest()
	currentdir := osutils.GetCWD()

	path := osutils.Path("../test_data/shell/echo")
	os.Chdir(path)
	as := assert.New(t)
	rootCmd.SetArgs([]string{"create", "--dry-run"})

	_, err := rootCmd.ExecuteC()
	as.NoError(err)

	os.Chdir(currentdir)
}

func TestCreateCommandExplicitPath(t *testing.T) {

	rootCmd, createOptions := setupCreateTest()

	as := assert.New(t)
	rootCmd.SetArgs([]string{"create", "--dry-run", "-f", osutils.Path("../test_data/shell/echo"), "-v", "0.0.1-snapshot"})

	_, err := rootCmd.ExecuteC()
	as.NoError(err)

	as.NotEmpty(createOptions.FilePath)
	as.Equal("../test_data/shell/echo", createOptions.FilePath)
}

func TestCreateCommandExplicitPathAndLang(t *testing.T) {

	rootCmd, createOptions := setupCreateTest()
	as := assert.New(t)
	rootCmd.SetArgs([]string{"create", "shell", "--dry-run", "-f", osutils.Path("../test_data/shell/echo"), "-v", "0.0.1-snapshot"})

	_, err := rootCmd.ExecuteC()
	as.NoError(err)

	as.NotEmpty(createOptions.FilePath)
	as.Equal("../test_data/shell/echo", createOptions.FilePath)
}

func TestCreateLanguageDoesNotMatchArtifact(t *testing.T) {

	rootCmd, _ := setupCreateTest()
	as := assert.New(t)
	rootCmd.SetArgs([]string{"create", "shell", "--dry-run", "-f", osutils.Path("../test_data/python/demo"), "-a", "demo.py"})

	_, err := rootCmd.ExecuteC()
	as.Error(err)
	as.Equal("language shell conflicts with artifact file extension demo.py", err.Error())
}

func TestCreatePythonCommand(t *testing.T) {
	rootCmd, createOptions := setupCreateTest()
	as := assert.New(t)
	rootCmd.SetArgs([]string{"create", "python", "--dry-run", "-f", osutils.Path("../test_data/python/demo"), "-v", "0.0.1-snapshot", "--handler", "process"})

	_, err := rootCmd.ExecuteC()
	as.NoError(err)
	as.Equal("process", createOptions.Handler)
}

func TestCreatePythonCommandWithDefaultHandler(t *testing.T) {

	rootCmd, createOptions := setupCreateTest()
	as := assert.New(t)
	rootCmd.SetArgs([]string{"create", "python", "--dry-run", "-f", osutils.Path("../test_data/python/demo"), "-v", "0.0.1-snapshot"})

	_, err := rootCmd.ExecuteC()
	as.NoError(err)
	as.Equal("demo", createOptions.Handler)
}

func TestInitCommandImplicitPath(t *testing.T) {
	rootCmd, _, _ := setupInitTest()
	as := assert.New(t)
	rootCmd.SetArgs([]string{"init", "--dry-run", "../test_data/shell/echo", "-v", "0.0.1-snapshot"})

	_, err := rootCmd.ExecuteC()
	as.NoError(err)

	//	as.NotEmpty(opts.InitOptions.FilePath)
	//	as.Equal("../test_data/shell/echo", opts.InitOptions.FilePath)
}

func TestInitCommandExplicitPath(t *testing.T) {
	rootCmd, _, _ := setupInitTest()
	as := assert.New(t)
	rootCmd.SetArgs([]string{"init", "--dry-run", "-f", "../test_data/shell/echo", "-v", "0.0.1-snapshot"})

	_, err := rootCmd.ExecuteC()
	as.NoError(err)

	//	as.NotEmpty(opts.InitOptions.FilePath)
	//	as.Equal("../test_data/shell/echo", opts.InitOptions.FilePath)
}

func TestInitCommandExplicitPathAndLang(t *testing.T) {
	rootCmd, _, _ := setupInitTest()
	as := assert.New(t)
	rootCmd.SetArgs([]string{"init", "shell", "--dry-run", "-f", "../test_data/shell/echo", "-v", "0.0.1-snapshot"})

	_, err := rootCmd.ExecuteC()
	as.NoError(err)

	//	as.NotEmpty(opts.InitOptions.FilePath)
	//	as.Equal("../test_data/shell/echo", opts.InitOptions.FilePath)
}

func TestInitJava(t *testing.T) {
	rootCmd, _, _ := setupInitTest()
	currentdir := osutils.GetCWD()
	path := osutils.Path("../test_data/java")
	os.Chdir(path)
	as := assert.New(t)
	//rootCmd.SetArgs([]string{"init", "java", "--dry-run", "-a", "target/dummy.jar", "--handler",  "function.Dummy", "-f","."})
	rootCmd.SetArgs([]string{"init", "java", "--dry-run", "-a", "target/dummy.jar", "--handler", "function.Dummy"})

	_, err := rootCmd.ExecuteC()
	fmt.Printf("%v\n", err)
	as.NoError(err)
	os.Chdir(currentdir)
}

func TestInitJavaWithVersion(t *testing.T) {

	rootCmd, _, _ := setupInitTest()

	currentdir := osutils.GetCWD()
	path := osutils.Path("../test_data/java")
	os.Chdir(path)
	as := assert.New(t)
	rootCmd.SetArgs([]string{"init", "java", "--dry-run", "-a", "target/upper-1.0.0.jar", "--handler", "function.Upper"})
	_, err := rootCmd.ExecuteC()
	as.NoError(err)
	os.Chdir(currentdir)
}

func TestCreateJavaWithVersion(t *testing.T) {

	rootCmd, _ := setupCreateTest()
	currentdir := osutils.GetCWD()
	path := osutils.Path("../test_data/java")
	os.Chdir(path)
	as := assert.New(t)
	rootCmd.SetArgs([]string{"create", "java", "--dry-run", "-a", "target/upper-1.0.0.jar", "--handler", "function.Upper"})
	_, err := rootCmd.ExecuteC()
	as.NoError(err)
	os.Chdir(currentdir)
}

func setupInitTest() (*cobra.Command, *options.InitOptions, map[string]*cobra.Command) {
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

	commands := make(map[string]*cobra.Command)
	registerCommand(commands, rootCmd)
	registerCommand(commands, initCmd)
	registerCommand(commands, initJavaCmd)
	registerCommand(commands, initNodeCmd)
	registerCommand(commands, initPythonCmd)
	registerCommand(commands, initShellCmd)
	registerCommand(commands, initGoCmd)

	return rootCmd, initOptions, commands
}

func registerCommand(commands map[string]*cobra.Command, command *cobra.Command) {
	commands[command.Name()] = command
}

func setupCreateTest() (*cobra.Command, *options.CreateOptions) {
	rootCmd, _, initCommands := setupInitTest()

	buildCmd, _ := Build()

	applyCmd, _ := Apply()

	createCmd, createOptions := Create(initCommands["init"], buildCmd, applyCmd)

	createNodeCmd, _ := CreateNode(initCommands["node"], buildCmd, applyCmd, createOptions)
	createJavaCmd, _ := CreateJava(initCommands["java"], buildCmd, applyCmd, createOptions)
	createPythonCmd, _ := CreatePython(initCommands["python"], buildCmd, applyCmd, createOptions)
	createShellCmd, _ := CreateShell(initCommands["shell"], buildCmd, applyCmd, createOptions)
	createGoCmd, _ := CreateGo(initCommands["go"], buildCmd, applyCmd, createOptions)

	createCmd.AddCommand(
		createNodeCmd,
		createJavaCmd,
		createPythonCmd,
		createShellCmd,
		createGoCmd,
	)

	return rootCmd, createOptions
}
