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
	"github.com/dturanski/riff-cli/pkg/osutils"
	"path/filepath"
	"os"
	"fmt"
)

func TestResolveDefaultFunctionResource(t *testing.T) {
	as := assert.New(t)
	currentDir := osutils.GetCWD()
	os.Chdir(osutils.Path("test_dir/python/demo"))
	options := InitOptions{functionPath: ""}
	functionPath, err := resolveFunctionPath(options, "py")
	if as.NoError(err) {
		absPath, _ := filepath.Abs(osutils.Path("demo.py"))
		as.Equal(absPath, functionPath)
	}
	os.Chdir(currentDir)
}

func TestResolveFunctionResourceFromFilePath(t *testing.T) {
	as := assert.New(t)
	options := InitOptions{functionPath: osutils.Path("test_dir/python/demo")}
	functionPath, err := resolveFunctionPath(options, "py")
	as.NoError(err)

	absPath, _ := filepath.Abs(osutils.Path("test_dir/python/demo/demo.py"))

	as.Equal(absPath, functionPath)
}

func TestResolveFunctionResourceFromFunctionFile(t *testing.T) {
	as := assert.New(t)
	options := InitOptions{functionPath: osutils.Path("test_dir/python/demo/demo.py")}
	functionPath, err := resolveFunctionPath(options, "py")
	as.NoError(err)

	absPath, _ := filepath.Abs(osutils.Path("test_dir/python/demo/demo.py"))

	as.Equal(absPath, functionPath)
}

func TestResolveFunctionResourceWithMultipleFilesPresent(t *testing.T) {
	as := assert.New(t)
	options := InitOptions{functionPath: osutils.Path("test_dir/python/multiple")}
	functionPath, err := resolveFunctionPath(options, "py")
	as.NoError(err)

	absPath, _ := filepath.Abs(osutils.Path("test_dir/python/multiple/multiple.py"))

	as.Equal(absPath, functionPath)
}

func TestResolveFunctionResourceFromArtifact(t *testing.T) {
	as := assert.New(t)
	options := InitOptions{functionPath: osutils.Path("test_dir/python/multiple"), artifact: "one.py"}
	functionPath, err := resolveFunctionPath(options, "py")
	as.NoError(err)

	absPath, _ := filepath.Abs(osutils.Path("test_dir/python/multiple/one.py"))

	as.Equal(absPath, functionPath)
}

func TestFunctionResourceDoesNotExist(t *testing.T) {
	as := assert.New(t)
	options := InitOptions{functionPath: osutils.Path("test_dir/python/demo")}
	functionPath, err := resolveFunctionPath(options, "js")
	as.Error(err)
	fmt.Println(functionPath)
}

func TestResolveFunctionResourceWithNoExtensionGiven(t *testing.T) {
	as := assert.New(t)
	options := InitOptions{functionPath: osutils.Path("test_dir/python/demo")}
	functionPath, err := resolveFunctionPath(options, "")
	if as.NoError(err) {
		absPath, _ := filepath.Abs(osutils.Path("test_dir/python/demo/demo.py"))
		as.Equal(absPath, functionPath)
	}
}

func TestFunctionResourceWithNoExtensionGivenDoesNotMatchFunctionName(t *testing.T) {
	as := assert.New(t)
	options := InitOptions{functionPath: osutils.Path("test_dir/python/demo"), functionName: "foo"}
	functionPath, err := resolveFunctionPath(options, "")
	fmt.Println(functionPath)
	as.Error(err)
}

func TestFunctionResourceWithNoExtensionGivenNotUnique(t *testing.T) {
	as := assert.New(t)
	options := InitOptions{functionPath: osutils.Path("test_dir/python/multiple"), functionName: "one"}
	_, err := resolveFunctionPath(options, "")
	as.Error(err)
	as.Contains(err.Error(),"function file is not unique")
}
