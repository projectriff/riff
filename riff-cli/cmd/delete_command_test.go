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
	"github.com/projectriff/riff/riff-cli/cmd/opts"
)

var getFunctionCount, deleteFunctionCount, deleteTopicCount, deleteResourceCount int

func mockKubeCtl() {
	fmt.Println("Initializing Stub KubeCtl")

	indexOf := func(s []string, e string) int {
		for i, a := range s {
			if a == e {
				return i
			}
		}
		return -1
	}

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
		ifunc := indexOf(cmdArgs,"function")
		itopic := indexOf(cmdArgs,"topic")
		iresource :=indexOf(cmdArgs,"-f")
		if cmdArgs[0] == "delete" && ifunc > 0  {
			deleteFunctionCount = deleteFunctionCount + 1
			return fmt.Sprintf("function %s deleted.", cmdArgs[ifunc + 1]), nil
		} else if cmdArgs[0] == "delete" && itopic > 0 {
			deleteTopicCount = deleteTopicCount + 1
			return fmt.Sprintf("topic %s deleted.", cmdArgs[itopic + 1]), nil
		} else if cmdArgs[0] == "delete" && iresource > 0 {
			deleteResourceCount = deleteResourceCount + 1
			return fmt.Sprintf("resources %s deleted.", cmdArgs[iresource + 1]), nil
	}
		return "",nil
	}
}

func TestMain(m *testing.M) {
	actualKubectlExecForString, actualKubectlExecForBytes := kubectl.EXEC_FOR_STRING, kubectl.EXEC_FOR_BYTES
	defer func() {
		kubectl.EXEC_FOR_STRING = actualKubectlExecForString
		kubectl.EXEC_FOR_BYTES = actualKubectlExecForBytes
	}()
	mockKubeCtl()
	retCode := m.Run()
	os.Exit(retCode)
}

func TestDeleteCommandImplicitPath(t *testing.T) {
	resetTestState()

	as := assert.New(t)
	rootCmd.SetArgs([]string{"delete", osutils.Path("../test_data/shell/echo")})

	_, err := rootCmd.ExecuteC()
	as.NoError(err)
	as.Equal("../test_data/shell/echo", opts.DeleteOptions.FilePath)
	as.Equal("", opts.DeleteOptions.Namespace)
	as.Equal(0, getFunctionCount)
	as.Equal(1, deleteFunctionCount)
	as.Equal(0, deleteTopicCount)
}

func TestDeleteCommandExplicitPath(t *testing.T) {
	resetTestState()
	as := assert.New(t)
	rootCmd.SetArgs([]string{"delete", "-f", osutils.Path("../test_data/shell/echo")})

	_, err := rootCmd.ExecuteC()
	as.NoError(err)
	as.Equal("../test_data/shell/echo", opts.DeleteOptions.FilePath)
	as.Equal("", opts.DeleteOptions.Namespace)
	as.Equal(0, getFunctionCount)
	as.Equal(1, deleteFunctionCount)
	as.Equal(0, deleteTopicCount)
	as.Equal(0, deleteResourceCount)
}

func TestDeleteCommandExplicitFile(t *testing.T) {
	resetTestState()
	as := assert.New(t)
	rootCmd.SetArgs([]string{"delete", "-f", osutils.Path("../test_data/shell/echo/echo-topics.yaml")})

	_, err := rootCmd.ExecuteC()
	as.NoError(err)
	as.Equal("../test_data/shell/echo/echo-topics.yaml", opts.DeleteOptions.FilePath)
	as.Equal("", opts.DeleteOptions.Namespace)
	as.Equal(0, getFunctionCount)
	as.Equal(1, deleteResourceCount)
	as.Equal(0, deleteTopicCount)
	as.Equal(0, deleteFunctionCount)
}

func TestDeleteCommandWithNameDoesNotExist(t *testing.T) {
	resetTestState()
	actualKubectlExecForBytes := kubectl.EXEC_FOR_BYTES
	defer func() {
		kubectl.EXEC_FOR_BYTES = actualKubectlExecForBytes
	}()
	// Just to avoid test dependence on kubectl
	kubectl.EXEC_FOR_BYTES = func(cmdArgs []string) ([]byte, error) {
		getFunctionCount = getFunctionCount + 1
		return ([]byte)("Mock: Error from server (NotFound): functions.projectriff.io square") , errors.New("Exit status1")
	}

	as := assert.New(t)
	rootCmd.SetArgs([]string{"delete", "--name", "square"})
	_, err := rootCmd.ExecuteC()
	as.Error(err)
	as.Equal("square", opts.DeleteOptions.FunctionName)
	as.Equal(1, getFunctionCount)
	as.Equal(0, deleteFunctionCount)
	as.Equal(0, deleteTopicCount)
	as.Equal(0, deleteResourceCount)

}

func TestDeleteCommandAllFlag(t *testing.T) {
	resetTestState()
	as := assert.New(t)
	rootCmd.SetArgs([]string{"delete", "-f", osutils.Path("../test_data/shell/echo"), "--all"})

	_, err := rootCmd.ExecuteC()
	as.NoError(err)
	as.Equal("../test_data/shell/echo", opts.DeleteOptions.FilePath)
	as.Equal(true, opts.DeleteOptions.All)
	as.Equal("", opts.DeleteOptions.Namespace)
	as.Equal(0, getFunctionCount)
	as.Equal(1, deleteResourceCount)
	as.Equal(0, deleteTopicCount)
	as.Equal(0, deleteFunctionCount)
}

func TestDeleteCommandFromCwdAllFlag(t *testing.T) {
	resetTestState()
	currentdir := osutils.GetCWD()
	defer func() { os.Chdir(currentdir) }()

	path := osutils.Path("../test_data/shell/echo")
	os.Chdir(path)
	as := assert.New(t)
	rootCmd.SetArgs([]string{"delete", "--all"})

	_, err := rootCmd.ExecuteC()
	as.NoError(err)
	as.Equal("", opts.DeleteOptions.FilePath)
	as.Equal(true, opts.DeleteOptions.All)
	as.Equal("", opts.DeleteOptions.Namespace)
	as.Equal(0, getFunctionCount)
	as.Equal(1, deleteResourceCount)
	as.Equal(0, deleteTopicCount)
	as.Equal(0, deleteFunctionCount)
}

func TestDeleteCommandFromCwdAllFlagNoResources(t *testing.T) {
	resetTestState()
	currentdir := osutils.GetCWD()
	defer func() { os.Chdir(currentdir) }()

	path := osutils.Path("../test_data/node/square")
	os.Chdir(path)
	as := assert.New(t)
	rootCmd.SetArgs([]string{"delete", "--all"})

	_, err := rootCmd.ExecuteC()
	as.NoError(err)
	as.Equal("", opts.DeleteOptions.FilePath)
	as.Equal(true, opts.DeleteOptions.All)
	as.Equal("", opts.DeleteOptions.Namespace)
	as.Equal(0, getFunctionCount)
	as.Equal(0, deleteResourceCount)
	as.Equal(0, deleteTopicCount)
	as.Equal(0, deleteFunctionCount)
}

func TestDeleteCommandWithFunctionName(t *testing.T) {
	resetTestState()

	as := assert.New(t)
	rootCmd.SetArgs([]string{"delete", "--all", "--name", "echo"})
	_, err := rootCmd.ExecuteC()
	as.NoError(err)
	as.Equal(1, getFunctionCount)
	as.Equal(1, deleteFunctionCount)
	as.Equal(2, deleteTopicCount)
	as.Equal(0, deleteResourceCount)
}

func TestDeleteCommandWithNamespace(t *testing.T) {
	resetTestState()
	as := assert.New(t)
	rootCmd.SetArgs([]string{"delete", "--dry-run", "--namespace", "test-test", "-f", osutils.Path("../test_data/shell/echo/")})

	_, err := rootCmd.ExecuteC()
	as.NoError(err)
	as.Equal("../test_data/shell/echo", opts.DeleteOptions.FilePath)
	as.Equal("test-test", opts.DeleteOptions.Namespace)
}

func resetTestState() {
	getFunctionCount = 0
	deleteFunctionCount = 0
	deleteTopicCount = 0
	deleteResourceCount = 0

	opts.DeleteOptions = options.DeleteOptions{}
	deleteCmd.ResetFlags()
	utils.CreateDeleteFlags(deleteCmd.Flags())
}
