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
	"github.com/projectriff/riff-cli/pkg/options"
)

func TestDeleteCommandImplicitPath(t *testing.T) {
	clearDeleteOptions()
	as := assert.New(t)
	rootCmd.SetArgs([]string{"delete", "--dry-run", osutils.Path("../test_data/shell/echo")})

	_, err := rootCmd.ExecuteC()
	as.NoError(err)
	as.Equal("../test_data/shell/echo", DeleteAllOptions.FunctionPath)
}

func TestDeleteCommandExplicitPath(t *testing.T) {
	clearDeleteOptions()
	as := assert.New(t)
	rootCmd.SetArgs([]string{"delete", "--dry-run", "-f", osutils.Path("../test_data/shell/echo")})

	_, err := rootCmd.ExecuteC()
	as.NoError(err)
	as.Equal("../test_data/shell/echo", DeleteAllOptions.FunctionPath)
}

func TestDeleteCommandExplicitFile(t *testing.T) {
	clearDeleteOptions()
	as := assert.New(t)
	rootCmd.SetArgs([]string{"delete", "--dry-run", "-f", osutils.Path("../test_data/shell/echo/echo-topics.yaml")})

	_, err := rootCmd.ExecuteC()
	as.NoError(err)
	as.Equal("../test_data/shell/echo/echo-topics.yaml", DeleteAllOptions.FunctionPath)
}

func TestDeleteCommandAllFlag(t *testing.T) {
	clearDeleteOptions()
	as := assert.New(t)
	rootCmd.SetArgs([]string{"delete", "--dry-run", "-f", osutils.Path("../test_data/shell/echo"), "--all"})

	_, err := rootCmd.ExecuteC()
	as.NoError(err)
	as.Equal("../test_data/shell/echo", DeleteAllOptions.FunctionPath)
	as.Equal(true, DeleteAllOptions.All)
}

func clearDeleteOptions() {
	DeleteAllOptions = options.DeleteAllOptions{}
}