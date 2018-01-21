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
)

const Ignore = false

func TestMain(m *testing.M) {
	var code int
	if !Ignore {
		code = m.Run()
	}
	os.Exit(code)
}

func TestCreateCommandImplicitPath(t *testing.T) {
	clearInitOptions()
	as := assert.New(t)
	//os.Chdir("test_dir/shell/echo")
	rootCmd.SetArgs([]string{"create", "--dry-run","test_dir/shell/echo","-v","0.0.1-snapshot"})

	_, err := rootCmd.ExecuteC()
	as.NoError(err)

	as.NotEmpty(initOptions.functionPath)
	as.Equal("test_dir/shell/echo",initOptions.functionPath)
	as.NoError(err)
}

func TestCreateCommandExplicitPath(t *testing.T) {
	clearInitOptions()
	as := assert.New(t)
	//os.Chdir("test_dir/shell/echo")
	rootCmd.SetArgs([]string{"create","--dry-run","-f","test_dir/shell/echo","-v","0.0.1-snapshot"})

	_, err := rootCmd.ExecuteC()
	as.NoError(err)

	as.NotEmpty(initOptions.functionPath)
	as.Equal("test_dir/shell/echo",initOptions.functionPath)
	as.NoError(err)
}

func TestCreateCommandExplicitPathAndLang(t *testing.T) {
	clearInitOptions()
	as := assert.New(t)
	//os.Chdir("test_dir/shell/echo")
	rootCmd.SetArgs([]string{"create", "shell","--dry-run","-f","test_dir/shell/echo","-v","0.0.1-snapshot"})

	_, err := rootCmd.ExecuteC()
	as.NoError(err)

	as.NotEmpty(initOptions.functionPath)
	as.Equal("test_dir/shell/echo",initOptions.functionPath)
	as.NoError(err)
}

func TestCreatePythonCommand(t *testing.T) {
	clearInitOptions()
	as := assert.New(t)
	//os.Chdir("test_dir/shell/echo")
	rootCmd.SetArgs([]string{"create", "python","--dry-run","-f","test_dir/python/demo","-v","0.0.1-snapshot","--handler","process"})

	_, err := rootCmd.ExecuteC()
	as.NoError(err)
}



func TestInitCommandImplicitPath(t *testing.T) {
	clearInitOptions()
	as := assert.New(t)
	//os.Chdir("test_dir/shell/echo")
	rootCmd.SetArgs([]string{"init", "test_dir/shell/echo","-v","0.0.1-snapshot"})

	_, err := rootCmd.ExecuteC()
	as.NoError(err)


	as.NotEmpty(initOptions.functionPath)
	as.Equal("test_dir/shell/echo",initOptions.functionPath)
	as.NoError(err)
}

func TestInitCommandExplicitPath(t *testing.T) {
	clearInitOptions()
	as := assert.New(t)
	//os.Chdir("test_dir/shell/echo")
	rootCmd.SetArgs([]string{"init","-f","test_dir/shell/echo","-v","0.0.1-snapshot"})

	_, err := rootCmd.ExecuteC()
	as.NoError(err)

	as.NotEmpty(initOptions.functionPath)
	as.Equal("test_dir/shell/echo",initOptions.functionPath)
	as.NoError(err)
}

func TestInitCommandExplicitPathAndLang(t *testing.T) {
	clearInitOptions()
	as := assert.New(t)
	//os.Chdir("test_dir/shell/echo")
	rootCmd.SetArgs([]string{"init", "shell","-f","test_dir/shell/echo","-v","0.0.1-snapshot"})

	_, err := rootCmd.ExecuteC()
	as.NoError(err)

	as.NotEmpty(initOptions.functionPath)
	as.Equal("test_dir/shell/echo",initOptions.functionPath)
	as.NoError(err)
}

func clearInitOptions(){
	initOptions.initialized = false
	initOptions.functionPath=""
}