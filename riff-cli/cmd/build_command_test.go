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

func TestBuildCommandImplicitPath(t *testing.T) {
	clearInitOptions()
	as := assert.New(t)
	rootCmd.SetArgs([]string{"build", "--dry-run", osutils.Path("../test_data/shell/echo"), "-v", "0.0.1-snapshot"})
	_, err := rootCmd.ExecuteC()
	as.Equal("echo", opts.CreateOptions.FunctionName)
	as.Equal("0.0.1-snapshot", opts.CreateOptions.Version)
	as.NoError(err)

}

func TestBuildCommandExplicitPath(t *testing.T) {
	clearInitOptions()
	as := assert.New(t)
	rootCmd.SetArgs([]string{"build", "--dry-run", "--push", "-f", osutils.Path("../test_data/shell/echo"), "-v", "0.0.2-snapshot"})

	_, err := rootCmd.ExecuteC()
	as.NoError(err)
	as.Equal("echo", opts.CreateOptions.FunctionName)
	as.Equal("0.0.2-snapshot", opts.CreateOptions.Version)
	as.True(opts.CreateOptions.Push)
}

func TestBuildCommandWithUserAccountAndVersion(t *testing.T) {
	clearInitOptions()
	as := assert.New(t)
	rootCmd.SetArgs([]string{"build", "--dry-run", "--push", "-f", osutils.Path("../test_data/shell/echo"), "-v", "0.0.1-snapshot","-u","projectriff"})

	_, err := rootCmd.ExecuteC()
	as.NoError(err)
	as.Equal("echo", opts.CreateOptions.FunctionName)
	as.Equal("0.0.1-snapshot", opts.CreateOptions.Version)
	as.Equal("projectriff", opts.CreateOptions.UserAccount)
	as.True(opts.CreateOptions.Push)
}

