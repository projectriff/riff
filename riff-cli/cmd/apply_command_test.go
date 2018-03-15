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
	"github.com/spf13/cobra"
)

func TestApplyCommandImplicitPath(t *testing.T) {
	as := assert.New(t)
	
	rootCmd, _, applyOptions := setupApplyTest()
	rootCmd.SetArgs([]string{"apply", "--dry-run", osutils.Path("../test_data/command/echo")})

	err := rootCmd.Execute()
	as.NoError(err)
	as.Equal("../test_data/command/echo", applyOptions.FilePath)
}

func TestApplyCommandExplicitPath(t *testing.T) {
	as := assert.New(t)

	rootCmd, _, applyOptions := setupApplyTest()
	rootCmd.SetArgs([]string{"apply", "--dry-run", "-f", osutils.Path("../test_data/command/echo")})

	_, err := rootCmd.ExecuteC()
	as.NoError(err)
	as.Equal("../test_data/command/echo", applyOptions.FilePath)
}

func TestApplyCommandDefaultNamespace(t *testing.T) {
	as := assert.New(t)
	rootCmd, _, applyOptions := setupApplyTest()
	rootCmd.SetArgs([]string{"apply", "--dry-run", "-f", osutils.Path("../test_data/command/echo")})

	_, err := rootCmd.ExecuteC()
	as.NoError(err)
	as.Equal("", applyOptions.Namespace)
}

func TestApplyCommandWithNamespace(t *testing.T) {

	as := assert.New(t)
	rootCmd, _, applyOptions := setupApplyTest()
	rootCmd.SetArgs([]string{"apply", "--dry-run", "--namespace", "test-test", "-f", osutils.Path("../test_data/command/echo")})

	_, err := rootCmd.ExecuteC()
	as.NoError(err)
	as.Equal("test-test", applyOptions.Namespace)
}

func setupApplyTest() (*cobra.Command, *cobra.Command, *ApplyOptions, ){
	root := Root()
	apply, applyOptions := Apply()
	root.AddCommand(apply)
	return root, apply, applyOptions
}