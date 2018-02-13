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
	"github.com/projectriff/riff-cli/pkg/osutils"
	"github.com/projectriff/riff-cli/cmd/opts"
)

func TestApplyCommandImplicitPath(t *testing.T) {
	clearAllOptions()
	as := assert.New(t)
	rootCmd.SetArgs([]string{"apply", "--dry-run", osutils.Path("../test_data/shell/echo")})

	_, err := rootCmd.ExecuteC()
	as.NoError(err)
	as.Equal("../test_data/shell/echo", opts.CreateOptions.FunctionPath)
}

func TestApplyCommandExplicitPath(t *testing.T) {
	clearAllOptions()
	as := assert.New(t)
	rootCmd.SetArgs([]string{"apply", "--dry-run", "-f", osutils.Path("../test_data/shell/echo")})

	_, err := rootCmd.ExecuteC()
	as.NoError(err)
	as.Equal("../test_data/shell/echo", opts.CreateOptions.FunctionPath)
}

func TestApplyCommandDefaultNamespace(t *testing.T) {
	clearAllOptions()
	as := assert.New(t)
	rootCmd.SetArgs([]string{"apply", "--dry-run", "-f", osutils.Path("../test_data/shell/echo")})

	_, err := rootCmd.ExecuteC()
	as.NoError(err)
	as.Equal("default", opts.CreateOptions.Namespace)
}

func TestApplyCommandWithNamespace(t *testing.T) {
	clearAllOptions()
	as := assert.New(t)
	rootCmd.SetArgs([]string{"apply", "--dry-run", "--namespace", "test-test", "-f", osutils.Path("../test_data/shell/echo")})

	_, err := rootCmd.ExecuteC()
	as.NoError(err)
	as.Equal("test-test", opts.CreateOptions.Namespace)
}
