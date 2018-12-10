/*
 * Copyright 2018 The original author or authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package commands

import (
	"fmt"

	"github.com/projectriff/riff/pkg/env"
	"github.com/spf13/cobra"
)

const (
	versionNumberOfArgs = iota
)

func Version() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version information about " + env.Cli.Name,
		Args:  cobra.ExactArgs(versionNumberOfArgs),
		RunE: func(cmd *cobra.Command, args []string) error {
			dirtyMsg := ""
			if env.Cli.GitDirty != "" {
				dirtyMsg = ", with local modifications"
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Version\n  %s cli: %s (%s%s)\n", env.Cli.Name, env.Cli.Version, env.Cli.GitSha, dirtyMsg)
			return nil
		},
	}
}
