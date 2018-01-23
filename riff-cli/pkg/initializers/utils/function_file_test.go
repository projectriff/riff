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

package utils

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"github.com/projectriff/riff-cli/pkg/osutils"
	"path/filepath"
	"os"
	"fmt"
	"github.com/projectriff/riff-cli/pkg/options"
)

var testDataRoot = "../../../test_data"

func TestResolveDefaultFunctionResource(t *testing.T) {
	as := assert.New(t)
	currentDir := osutils.GetCWD()
	os.Chdir(osutils.Path(testDataRoot + "/python/demo"))
	opts := options.InitOptions{FunctionPath: osutils.GetCWD()}
	options.ValidateAndCleanInitOptions(&opts)
	functionPath, err := ResolveFunctionFile(opts, "python","py")
	if as.NoError(err) {
		absPath, _ := filepath.Abs(osutils.Path("demo.py"))
		as.Equal(absPath, functionPath)
	}
	os.Chdir(currentDir)
}

func TestResolveFunctionResourceFromFilePath(t *testing.T) {
	as := assert.New(t)
	opts := options.InitOptions{FunctionPath: osutils.Path(testDataRoot + "/python/demo")}
	options.ValidateAndCleanInitOptions(&opts)
	functionPath, err := ResolveFunctionFile(opts, "python","py")
	as.NoError(err)

	absPath, _ := filepath.Abs(osutils.Path(testDataRoot + "/python/demo/demo.py"))

	as.Equal(absPath, functionPath)
}

func TestResolveFunctionResourceFromFunctionFile(t *testing.T) {
	as := assert.New(t)
	opts := options.InitOptions{FunctionPath: osutils.Path(testDataRoot + "/python/demo/demo.py")}
	options.ValidateAndCleanInitOptions(&opts)
	functionPath, err := ResolveFunctionFile(opts, "python", "py")
	as.NoError(err)

	absPath, _ := filepath.Abs(osutils.Path(testDataRoot + "/python/demo/demo.py"))

	as.Equal(absPath, functionPath)
}

func TestResolveFunctionResourceWithMultipleFilesPresent(t *testing.T) {
	as := assert.New(t)
	opts := options.InitOptions{FunctionPath: osutils.Path(testDataRoot + "/python/multiple")}
	options.ValidateAndCleanInitOptions(&opts)
	functionPath, err := ResolveFunctionFile(opts, "python","py")
	as.NoError(err)

	absPath, _ := filepath.Abs(osutils.Path(testDataRoot + "/python/multiple/multiple.py"))

	as.Equal(absPath, functionPath)
}

func TestResolveFunctionResourceFromArtifact(t *testing.T) {
	as := assert.New(t)
	opts := options.InitOptions{FunctionPath: osutils.Path(testDataRoot + "/python/multiple"), Artifact: "one.py"}
	options.ValidateAndCleanInitOptions(&opts)
	functionPath, err := ResolveFunctionFile(opts, "python","py")
	as.NoError(err)

	absPath, _ := filepath.Abs(osutils.Path(testDataRoot + "/python/multiple/one.py"))

	as.Equal(absPath, functionPath)
}

func TestFunctionResourceDoesNotExist(t *testing.T) {
	as := assert.New(t)
	opts := options.InitOptions{FunctionPath: osutils.Path(testDataRoot + "/python/demo")}
	options.ValidateAndCleanInitOptions(&opts)
	functionPath, err := ResolveFunctionFile(opts, "node","js")
	as.Error(err)
	fmt.Println(functionPath)
}

func TestResolveFunctionResourceWithNoExtensionGiven(t *testing.T) {
	as := assert.New(t)
	opts := options.InitOptions{FunctionPath: osutils.Path(testDataRoot + "/python/demo")}
	options.ValidateAndCleanInitOptions(&opts)
	functionPath, err := ResolveFunctionFile(opts, "","")
	if as.NoError(err) {
		absPath, _ := filepath.Abs(osutils.Path(testDataRoot + "/python/demo/demo.py"))
		as.Equal(absPath, functionPath)
	}
}

func TestFunctionResourceWithNoExtensionGivenDoesNotMatchFunctionName(t *testing.T) {
	as := assert.New(t)
	opts := options.InitOptions{FunctionPath: osutils.Path(testDataRoot + "/python/demo"), FunctionName: "foo"}
	options.ValidateAndCleanInitOptions(&opts)
	functionPath, err := ResolveFunctionFile(opts, "","")
	fmt.Println(functionPath)
	as.Error(err)
}

func TestFunctionResourceWithNoExtensionGivenNotUnique(t *testing.T) {
	as := assert.New(t)
	opts := options.InitOptions{FunctionPath: osutils.Path(testDataRoot + "/python/multiple"), FunctionName: "one"}
	options.ValidateAndCleanInitOptions(&opts)
	_, err := ResolveFunctionFile(opts, "","")
	as.Error(err)
	as.Contains(err.Error(),"function file is not unique")
}
