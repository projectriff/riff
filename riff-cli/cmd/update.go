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
	"github.com/spf13/cobra"
	"github.com/projectriff/riff/riff-cli/cmd/utils"
)

func Update(buildCmd *cobra.Command, applyCmd *cobra.Command) *cobra.Command {

	updateChainCmd := utils.CommandChain(buildCmd, applyCmd)

	updateChainCmd.Use = "update"
	updateChainCmd.Short = "Update a function (equivalent to build, apply)"
	updateChainCmd.Long = `Build the function based on the code available in the path directory, using the name and version specified 
for the image that is built. Then Apply the resource definition[s] included in the path.`
	updateChainCmd.Example = `  riff update -n <name> -v <version> -f <path> [--push]`

	return updateChainCmd
}
