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
	"os"
	"github.com/projectriff/riff/riff-cli/pkg/osutils"
	"github.com/spf13/cobra"
	"github.com/projectriff/riff/riff-cli/pkg/options"
)

func TestCreateCommandImplicitPath(t *testing.T) {
	rootCmd, initOptions, _, _:= setupCreateTest()
	as := assert.New(t)
	rootCmd.SetArgs([]string{"create", "--dry-run", "../test_data/shell/echo", "-v", "0.0.1-snapshot"})

	_, err := rootCmd.ExecuteC()
	as.NoError(err)

	as.NotEmpty(initOptions.FilePath)
	as.NotEmpty(initOptions.UserAccount)
	as.Equal("../test_data/shell/echo", initOptions.FilePath)
}

func TestCreateCommandWithUser(t *testing.T) {
	rootCmd, initOptions, _, _:= setupCreateTest()
	as := assert.New(t)
	rootCmd.SetArgs([]string{"create", "--dry-run", "../test_data/shell/echo", "-u", "me"})

	_, err := rootCmd.ExecuteC()
	as.NoError(err)

	as.NotEmpty(initOptions.FilePath)
	as.Equal("me", initOptions.UserAccount)
	as.Equal("../test_data/shell/echo", initOptions.FilePath)
}

func TestCreateCommandFromCWD(t *testing.T) {
	rootCmd, _, _, _:= setupCreateTest()
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
	rootCmd, initOptions, _, _:= setupCreateTest()
	as := assert.New(t)
	rootCmd.SetArgs([]string{"create", "--dry-run", "-f", osutils.Path("../test_data/shell/echo"), "-v", "0.0.1-snapshot"})

	_, err := rootCmd.ExecuteC()
	as.NoError(err)

	as.NotEmpty(initOptions.FilePath)
	as.NotEmpty(initOptions.UserAccount)
	as.Equal("../test_data/shell/echo", initOptions.FilePath)
}

func TestCreateCommandWithUser(t *testing.T) {
	rootCmd, initOptions, _, _:= setupCreateTest()
	as := assert.New(t)
	rootCmd.SetArgs([]string{"create", "--dry-run", "../test_data/shell/echo", "-u", "me"})

	_, err := rootCmd.ExecuteC()
	as.NoError(err)

	as.NotEmpty(initOptions.FilePath)
	as.Equal("me", initOptions.UserAccount)
	as.Equal("../test_data/shell/echo", initOptions.FilePath)
}

func TestCreateCommandExplicitPathAndLang(t *testing.T) {
	rootCmd, initOptions, _, _:= setupCreateTest()
	as := assert.New(t)
	rootCmd.SetArgs([]string{"create", "shell", "--dry-run", "-f", osutils.Path("../test_data/shell/echo"), "-v", "0.0.1-snapshot"})

	_, err := rootCmd.ExecuteC()
	as.NoError(err)

	as.NotEmpty(initOptions.FilePath)
	as.Equal("../test_data/shell/echo", initOptions.FilePath)
}

func TestCreateLanguageDoesNotMatchArtifact(t *testing.T) {
	rootCmd, _, _, _:= setupCreateTest()
	as := assert.New(t)
	rootCmd.SetArgs([]string{"create", "shell", "--dry-run", "-f", osutils.Path("../test_data/python/demo"), "-a", "demo.py"})

	_, err := rootCmd.ExecuteC()
	as.Error(err)
	as.Equal("language shell conflicts with artifact file extension demo.py", err.Error())
}

func TestCreatePythonCommand(t *testing.T) {
	rootCmd, initOptions, _, _:= setupCreateTest()
	as := assert.New(t)
	rootCmd.SetArgs([]string{"create", "python", "--dry-run", "-f", osutils.Path("../test_data/python/demo"), "-v", "0.0.1-snapshot", "--handler", "process"})

	_, err := rootCmd.ExecuteC()
	as.NoError(err)
	as.NotEmpty(initOptions.UserAccount)
	as.Equal("process", initOptions.Handler)
}

func TestCreatePythonCommandWithDefaultHandler(t *testing.T) {
	rootCmd, initOptions, _, _:= setupCreateTest()
	as := assert.New(t)
	rootCmd.SetArgs([]string{"create", "python", "--dry-run", "-f", osutils.Path("../test_data/python/demo"), "-v", "0.0.1-snapshot"})

	_, err := rootCmd.ExecuteC()
	as.NoError(err)
	as.Equal("demo", initOptions.Handler)
}

func TestCreateJavaWithVersion(t *testing.T) {
	rootCmd, initOptions, _, _:= setupCreateTest()
	currentdir := osutils.GetCWD()
	path := osutils.Path("../test_data/java")
	os.Chdir(path)
	as := assert.New(t)
	rootCmd.SetArgs([]string{"create", "java", "--dry-run", "-a", "target/upper-1.0.0.jar", "--handler", "function.Upper"})
	_, err := rootCmd.ExecuteC()
	as.NoError(err)
	as.NotEmpty(initOptions.UserAccount)
	os.Chdir(currentdir)
}

func setupCreateTest() (*cobra.Command, *options.InitOptions, *BuildOptions, *ApplyOptions) {
	rootCmd, initOptions, initCommands := setupInitTest()

	buildCmd, buildOptions := Build()

	applyCmd, applyOptions := Apply()

	createCmd := Create(initCommands["init"], buildCmd, applyCmd)

	createNodeCmd := CreateNode(initCommands["node"], buildCmd, applyCmd)
	createJavaCmd := CreateJava(initCommands["java"], buildCmd, applyCmd)
	createPythonCmd := CreatePython(initCommands["python"], buildCmd, applyCmd)
	createShellCmd := CreateShell(initCommands["shell"], buildCmd, applyCmd)
	createGoCmd := CreateGo(initCommands["go"], buildCmd, applyCmd)

	createCmd.AddCommand(
		createNodeCmd,
		createJavaCmd,
		createPythonCmd,
		createShellCmd,
		createGoCmd,
	)

	rootCmd.AddCommand(
		createCmd,
	)


	return rootCmd, initOptions, buildOptions, applyOptions
}
