/*
 * Copyright 2018 the original author or authors.
 *
 *   Licensed under the Apache License, Version 2.0 (the "License");
 *   you may not use this file except in compliance with the License.
 *   You may obtain a copy of the License at
 *
 *        https://www.apache.org/licenses/LICENSE-2.0
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
	"github.com/projectriff/riff/riff-cli/pkg/docker"
	"github.com/projectriff/riff/riff-cli/pkg/kubectl"
)

func TestUpdateCommandImplicitPath(t *testing.T) {
	rootCmd, buildOptions, applyOptions := setupUpdateTest()
	as := assert.New(t)
	rootCmd.SetArgs([]string{"update", "--dry-run", osutils.Path("../test_data/command/echo")})

	_, err := rootCmd.ExecuteC()
	as.NoError(err)
	as.Equal("../test_data/command/echo", buildOptions.FilePath)
	as.Equal("../test_data/command/echo", applyOptions.FilePath)
}

func TestUpdateCommandExplicitPath(t *testing.T) {
	rootCmd, buildOptions, applyOptions := setupUpdateTest()
	as := assert.New(t)
	rootCmd.SetArgs([]string{"update", "--dry-run", "-f", osutils.Path("../test_data/command/echo")})

	_, err := rootCmd.ExecuteC()
	as.NoError(err)
	as.Equal("../test_data/command/echo", buildOptions.FilePath)
	as.Equal("../test_data/command/echo", applyOptions.FilePath)
}

func setupUpdateTest() (*cobra.Command, *BuildOptions, *ApplyOptions) {
	rootCmd := Root()
	buildCmd, buildOptions := Build(docker.RealDocker(), docker.DryRunDocker())

	applyCmd, applyOptions := Apply(kubectl.RealKubeCtl(), kubectl.DryRunKubeCtl())

	update := Update(buildCmd, applyCmd)

	rootCmd.AddCommand(update,buildCmd,applyCmd)

	return rootCmd, buildOptions, applyOptions
}
