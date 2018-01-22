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
)

func TestCreateCommandImplicitPath(t *testing.T) {
	clearInitOptions()
	as := assert.New(t)
	rootCmd.SetArgs([]string{"create", "--dry-run", "test_dir/shell/echo", "-v", "0.0.1-snapshot"})

	_, err := rootCmd.ExecuteC()
	as.NoError(err)

	as.NotEmpty(initOptions.FunctionPath)
	as.Equal("test_dir/shell/echo", initOptions.FunctionPath)

}

func TestCreateCommandExplicitPath(t *testing.T) {
	clearInitOptions()
	as := assert.New(t)
	rootCmd.SetArgs([]string{"create", "--dry-run", "-f", "test_dir/shell/echo", "-v", "0.0.1-snapshot"})

	_, err := rootCmd.ExecuteC()
	as.NoError(err)

	as.NotEmpty(initOptions.FunctionPath)
	as.Equal("test_dir/shell/echo", initOptions.FunctionPath)
}

func TestCreateCommandExplicitPathAndLang(t *testing.T) {
	clearInitOptions()
	as := assert.New(t)
	rootCmd.SetArgs([]string{"create", "shell", "--dry-run", "-f", "test_dir/shell/echo", "-v", "0.0.1-snapshot"})

	_, err := rootCmd.ExecuteC()
	as.NoError(err)

	as.NotEmpty(initOptions.FunctionPath)
	as.Equal("test_dir/shell/echo", initOptions.FunctionPath)
}

func TestCreateLanguageDoesNotMatchArtifact(t *testing.T) {
	clearInitOptions()
	as := assert.New(t)
	rootCmd.SetArgs([]string{"create", "shell", "--dry-run", "-f", "test_dir/python/demo", "-a","demo.py"})

	_, err := rootCmd.ExecuteC()
	as.Error(err)
}

func TestCreatePythonCommand(t *testing.T) {
	clearInitOptions()
	as := assert.New(t)
	rootCmd.SetArgs([]string{"create", "python", "--dry-run", "-f", "test_dir/python/demo", "-v", "0.0.1-snapshot", "--handler", "process"})

	_, err := rootCmd.ExecuteC()
	as.NoError(err)
	as.Equal("process",handler)
}

func TestInitCommandImplicitPath(t *testing.T) {
	clearInitOptions()
	as := assert.New(t)
	rootCmd.SetArgs([]string{"init", "--dry-run", "test_dir/shell/echo", "-v", "0.0.1-snapshot"})

	_, err := rootCmd.ExecuteC()
	as.NoError(err)

	as.NotEmpty(initOptions.FunctionPath)
	as.Equal("test_dir/shell/echo", initOptions.FunctionPath)
}

func TestInitCommandExplicitPath(t *testing.T) {
	clearInitOptions()
	as := assert.New(t)
	rootCmd.SetArgs([]string{"init", "--dry-run", "-f", "test_dir/shell/echo", "-v", "0.0.1-snapshot"})

	_, err := rootCmd.ExecuteC()
	as.NoError(err)

	as.NotEmpty(initOptions.FunctionPath)
	as.Equal("test_dir/shell/echo", initOptions.FunctionPath)
}

func TestInitCommandExplicitPathAndLang(t *testing.T) {
	clearInitOptions()
	as := assert.New(t)
	rootCmd.SetArgs([]string{"init", "shell", "--dry-run", "-f", "test_dir/shell/echo", "-v", "0.0.1-snapshot"})

	_, err := rootCmd.ExecuteC()
	as.NoError(err)

	as.NotEmpty(initOptions.FunctionPath)
	as.Equal("test_dir/shell/echo", initOptions.FunctionPath)
}

func clearInitOptions() {
	initOptions = options.InitOptions{}
	createOptions = options.CreateOptions{}
}
