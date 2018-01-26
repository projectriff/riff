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
	"github.com/projectriff/riff-cli/pkg/options"

	"os"
	"github.com/projectriff/riff-cli/pkg/osutils"
	"github.com/projectriff/riff-cli/cmd/opts"
)

func TestCreateCommandImplicitPath(t *testing.T) {
	clearAllOptions()
	as := assert.New(t)
	rootCmd.SetArgs([]string{"create", "--dry-run", "../test_data/shell/echo", "-v", "0.0.1-snapshot"})

	_, err := rootCmd.ExecuteC()
	as.NoError(err)

	as.NotEmpty(opts.CreateOptions.FunctionPath)
	as.Equal("../test_data/shell/echo", opts.CreateOptions.FunctionPath)

}

func TestCreateCommandFromCWD(t *testing.T) {
	clearAllOptions()
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
	clearAllOptions()
	as := assert.New(t)
	rootCmd.SetArgs([]string{"create", "--dry-run", "-f", osutils.Path("../test_data/shell/echo"), "-v", "0.0.1-snapshot"})

	_, err := rootCmd.ExecuteC()
	as.NoError(err)

	as.NotEmpty(opts.CreateOptions.FunctionPath)
	as.Equal("../test_data/shell/echo", opts.CreateOptions.FunctionPath)
}

func TestCreateCommandExplicitPathAndLang(t *testing.T) {
	clearAllOptions()
	as := assert.New(t)
	rootCmd.SetArgs([]string{"create", "shell", "--dry-run", "-f", osutils.Path("../test_data/shell/echo"), "-v", "0.0.1-snapshot"})

	_, err := rootCmd.ExecuteC()
	as.NoError(err)

	as.NotEmpty(opts.CreateOptions.FunctionPath)
	as.Equal("../test_data/shell/echo", opts.CreateOptions.FunctionPath)
}

func TestCreateLanguageDoesNotMatchArtifact(t *testing.T) {
	clearAllOptions()
	as := assert.New(t)
	rootCmd.SetArgs([]string{"create", "shell", "--dry-run", "-f", osutils.Path("../test_data/python/demo"), "-a","demo.py"})

	_, err := rootCmd.ExecuteC()
	as.Error(err)
}

func TestCreatePythonCommand(t *testing.T) {
	clearAllOptions()
	as := assert.New(t)
	rootCmd.SetArgs([]string{"create", "python", "--dry-run", "-f", osutils.Path("../test_data/python/demo"), "-v", "0.0.1-snapshot", "--handler", "process"})

	_, err := rootCmd.ExecuteC()
	as.NoError(err)
	as.Equal("process", opts.Handler)
}

func TestInitCommandImplicitPath(t *testing.T) {
	clearAllOptions()
	as := assert.New(t)
	rootCmd.SetArgs([]string{"init", "--dry-run", "../test_data/shell/echo", "-v", "0.0.1-snapshot"})

	_, err := rootCmd.ExecuteC()
	as.NoError(err)

	as.NotEmpty(opts.InitOptions.FunctionPath)
	as.Equal("../test_data/shell/echo", opts.InitOptions.FunctionPath)
}

func TestInitCommandExplicitPath(t *testing.T) {
	clearAllOptions()
	as := assert.New(t)
	rootCmd.SetArgs([]string{"init", "--dry-run", "-f", "../test_data/shell/echo", "-v", "0.0.1-snapshot"})

	_, err := rootCmd.ExecuteC()
	as.NoError(err)

	as.NotEmpty(opts.InitOptions.FunctionPath)
	as.Equal("../test_data/shell/echo", opts.InitOptions.FunctionPath)
}

func TestInitCommandExplicitPathAndLang(t *testing.T) {
	clearAllOptions()
	as := assert.New(t)
	rootCmd.SetArgs([]string{"init", "shell", "--dry-run", "-f", "../test_data/shell/echo", "-v", "0.0.1-snapshot"})

	_, err := rootCmd.ExecuteC()
	as.NoError(err)

	as.NotEmpty(opts.InitOptions.FunctionPath)
	as.Equal("../test_data/shell/echo", opts.InitOptions.FunctionPath)
}

func clearAllOptions() {
	opts.InitOptions = options.InitOptions{}
	opts.CreateOptions = options.CreateOptions{}
	opts.AllOptions = options.AllOptions{}
	opts.Handler = ""
}
