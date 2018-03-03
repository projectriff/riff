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
	"github.com/projectriff/riff/riff-cli/pkg/osutils"
	"github.com/projectriff/riff/riff-cli/pkg/options"
	"os"
	"github.com/projectriff/riff/riff-cli/cmd/utils"
	"fmt"
	"github.com/projectriff/riff/riff-cli/pkg/kubectl"
	"errors"
)

func TestDeleteCommandImplicitPath(t *testing.T) {
	clearDeleteOptions()
	as := assert.New(t)
	rootCmd.SetArgs([]string{"delete", "--dry-run", osutils.Path("../test_data/shell/echo")})

	_, err := rootCmd.ExecuteC()
	as.NoError(err)
	as.Equal("../test_data/shell/echo", DeleteAllOptions.FilePath)
	as.Equal("default", DeleteAllOptions.Namespace)
}

func TestDeleteCommandExplicitPath(t *testing.T) {
	clearDeleteOptions()
	as := assert.New(t)
	rootCmd.SetArgs([]string{"delete", "--dry-run", "-f", osutils.Path("../test_data/shell/echo")})

	_, err := rootCmd.ExecuteC()
	as.NoError(err)
	as.Equal("../test_data/shell/echo", DeleteAllOptions.FilePath)
	as.Equal("default", DeleteAllOptions.Namespace)
}

func TestDeleteCommandExplicitFile(t *testing.T) {
	clearDeleteOptions()
	as := assert.New(t)
	rootCmd.SetArgs([]string{"delete", "--dry-run", "-f", osutils.Path("../test_data/shell/echo/echo-topics.yaml")})

	_, err := rootCmd.ExecuteC()
	as.NoError(err)
	as.Equal("../test_data/shell/echo/echo-topics.yaml", DeleteAllOptions.FilePath)
	as.Equal("default", DeleteAllOptions.Namespace)
}

func TestDeleteCommandWithName(t *testing.T) {
	clearDeleteOptions()
	actualKubectlExecForBytes:= kubectl.EXEC_FOR_BYTES
	defer func() {
		kubectl.EXEC_FOR_BYTES = actualKubectlExecForBytes
	}()

	// Just to avoid test dependence on kubectl
	kubectl.EXEC_FOR_BYTES = func(cmdArgs []string) ([]byte, error) {
		return ([]byte)("Mock: Error from server (NotFound): functions.projectriff.io square") , errors.New("Exit status1")
	}

	as := assert.New(t)
	rootCmd.SetArgs([]string{"delete", "--dry-run", "--name", "square"})
	_, err := rootCmd.ExecuteC()
	as.Error(err)
	as.Equal("square", DeleteAllOptions.FunctionName)

}

func TestDeleteCommandAllFlag(t *testing.T) {
	clearDeleteOptions()
	as := assert.New(t)
	rootCmd.SetArgs([]string{"delete", "--dry-run", "-f", osutils.Path("../test_data/shell/echo"), "--all"})

	_, err := rootCmd.ExecuteC()
	as.NoError(err)
	as.Equal("../test_data/shell/echo", DeleteAllOptions.FilePath)
	as.Equal(true, DeleteAllOptions.All)
	as.Equal("default", DeleteAllOptions.Namespace)
}

func TestDeleteCommandFromCwdAllFlag(t *testing.T) {
	clearDeleteOptions()
	currentdir := osutils.GetCWD()
	defer func() { os.Chdir(currentdir) }()

	path := osutils.Path("../test_data/shell/echo")
	os.Chdir(path)
	as := assert.New(t)
	rootCmd.SetArgs([]string{"delete", "--dry-run", "--all"})

	_, err := rootCmd.ExecuteC()
	as.NoError(err)
	as.Equal("", DeleteAllOptions.FilePath)
	as.Equal(true, DeleteAllOptions.All)
	as.Equal("default", DeleteAllOptions.Namespace)

}

func TestDeleteCommandWithFunctionName(t *testing.T) {
	clearDeleteOptions()
	actualKubectlExecForString, actualKubectlExecForBytes := kubectl.EXEC_FOR_STRING, kubectl.EXEC_FOR_BYTES
	defer func() {
		kubectl.EXEC_FOR_STRING = actualKubectlExecForString
		kubectl.EXEC_FOR_BYTES = actualKubectlExecForBytes
	}()

	getFunctionCount := 0;
	deleteFunctionCount := 0;
	deleteTopicCount := 0;
	kubectl.EXEC_FOR_BYTES = func(cmdArgs []string) ([]byte, error) {

		response := ([]byte)(
			`{
				"apiVersion": "projectriff.io/v1",
				"kind": "Function",
				"metadata": {},
				"spec": {
					"container": {
					"image": "test/echo:0.0.1"
					},
					"input": "myInputTopic",
					"output": "myOutputTopic",
					"protocol": "grpc"
				}
			}`)
		getFunctionCount = getFunctionCount + 1;
		return response, nil
	}

	kubectl.EXEC_FOR_STRING = func(cmdArgs []string) (string, error) {
		if (cmdArgs[1] == "function") {
			deleteFunctionCount = deleteFunctionCount + 1
			return fmt.Sprintf("function %s deleted.", cmdArgs[2]), nil
		} else if (cmdArgs[1] == "topic") {
			deleteTopicCount = deleteTopicCount + 1
			return fmt.Sprintf("topic %s deleted.", cmdArgs[2]), nil
		}
		return "",nil
	}

	as := assert.New(t)
	rootCmd.SetArgs([]string{"delete", "--all", "--name", "echo"})
	_, err := rootCmd.ExecuteC()
	as.NoError(err)
	as.Equal(1, getFunctionCount)
	as.Equal(1, deleteFunctionCount)
	as.Equal(2, deleteTopicCount)
}

func TestDeleteCommandWithNamespace(t *testing.T) {
	clearDeleteOptions()
	as := assert.New(t)
	rootCmd.SetArgs([]string{"delete", "--dry-run", "--namespace", "test-test", "-f", osutils.Path("../test_data/shell/echo/")})

	_, err := rootCmd.ExecuteC()
	as.NoError(err)
	as.Equal("../test_data/shell/echo", DeleteAllOptions.FilePath)
	as.Equal("test-test", DeleteAllOptions.Namespace)
}

func clearDeleteOptions() {
	DeleteAllOptions = options.DeleteAllOptions{}
	deleteCmd.ResetFlags()
	utils.CreateDeleteFlags(deleteCmd.Flags())
}
