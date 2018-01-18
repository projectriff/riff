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
	"fmt"
	"os"
	"path/filepath"
)

func TestValidateDefaultFunctionResources(t *testing.T) {
	filePath,_ := initCmd.PersistentFlags().GetString("filepath")
	as := assert.New(t)
	options:= InitOptions{functionPath: filePath, functionName:"foo"}
	as.NoError(validateAndCleanInitOptions(&options))
}

func TestValidateCleansFunctionResources(t *testing.T) {
	filePath := filepath.Join("test_dir","python","demo") + string(os.PathSeparator) + string(os.PathSeparator)
	artifact := "demo.py"
	as := assert.New(t)
	options:= InitOptions{functionPath: filePath, artifact:artifact}
	as.NoError(validateAndCleanInitOptions(&options))
	as.Equal(osutils.Path("test_dir/python/demo"),options.functionPath)
}

func TestValidateFunctionResourceDoesNotExist(t *testing.T) {
	filePath := osutils.Path("does/not/exist")
	as := assert.New(t)
	options:= InitOptions{functionPath: filePath}
	err := validateAndCleanInitOptions(&options)
	as.Error(err)
	as.Equal(fmt.Sprintf("filepath %s does not exist", filePath),err.Error())
}

func TestValidateArtifactIsRegularFile(t *testing.T) {
	filePath := osutils.Path("test_dir/python")
	as := assert.New(t)
	options:= InitOptions{functionPath: filePath, artifact: "demo"}
	err := validateAndCleanInitOptions(&options)
	as.Error(err)
	as.Contains(err.Error(), "must be a regular file")
}

func TestValidateArtifactIsInSubDirectory(t *testing.T) {
	filePath := osutils.Path("test_dir")
	as := assert.New(t)
	options:= InitOptions{functionPath: filePath, artifact: "python/demo/demo.py"}
	err := validateAndCleanInitOptions(&options)
	as.NoError(err)
}

func TestArtifactCannotBeExternalToFilePath(t *testing.T) {
	currentDir := osutils.GetCWD()
	os.Chdir(osutils.Path("test_dir/python/demo"))
	filePath := osutils.GetCWD()
	as := assert.New(t)
	options:= InitOptions{functionPath: filePath, artifact: osutils.Path("../multiple/one.py")}
	err := validateAndCleanInitOptions(&options)
	as.Error(err)
	as.Contains(err.Error(), "cannot be external to filepath",)
	os.Chdir(currentDir)
}

func TestArtifactRelativeToFilePath(t *testing.T) {
	filePath := osutils.Path("test_dir/python/demo")
	artifact := "demo.py"
	as := assert.New(t)
	options:= InitOptions{functionPath: filePath, artifact: artifact}
	err := validateAndCleanInitOptions(&options)
	as.NoError(err)
}


func TestAbsoluteArtifactPathConflctsFilePath(t *testing.T) {
	filePath := osutils.Path("test_dir/python/multiple/one.py")
	artifact :=  osutils.Path("two.py")
	as := assert.New(t)

	options:= InitOptions{functionPath: filePath, artifact: artifact}
	err := validateAndCleanInitOptions(&options)
	as.Contains(err.Error(), "conflicts with filepath")
	as.Error(err)
}

func TestInvalidProtocol(t *testing.T) {
	as := assert.New(t)
	options:= InitOptions{protocol:"grpz"}
	err := validateAndCleanInitOptions(&options)
	as.Error(err)
	as.Contains(err.Error(),"unsupported")
}

func TestCleanedProtocol(t *testing.T) {
	as := assert.New(t)
	options:= InitOptions{protocol:"gRPC"}
	err := validateAndCleanInitOptions(&options)
	as.NoError(err)
	as.Equal("grpc",options.protocol)
}

