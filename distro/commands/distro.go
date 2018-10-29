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
 *
 */

package commands

import (
	"github.com/projectriff/riff/cmd/commands"
	"github.com/projectriff/riff/pkg/core"
	"github.com/projectriff/riff/pkg/env"
	"github.com/spf13/cobra"
)

func DistroCreate(c *core.ImageClient) *cobra.Command {
	options := core.CreateDistroOptions{}

	command := &cobra.Command{
		Use:   "create",
		Short: "Create a " + env.Cli.Name + " distribution.",
		Long: "Create a " + env.Cli.Name + " distribution archive file (.tgz) from a given manifest.\n\n" +
			"If the output path is that of an existing directory, the file \"distro.tgz\" will be written in that " +
			"directory. Otherwise, the file will be written at the output path.",
		Example: "  " + env.Cli.Name + " create --output=./my-distro.tgz",
		Args:    cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			err := (*c).CreateDistro(options)
			if err != nil {
				return err
			}

			commands.PrintSuccessfulCompletion(cmd)
			return nil
		},
	}

	command.Flags().StringVarP(&options.Manifest, "manifest", "m", "stable", "manifest for the download; can be a named manifest (stable or latest) or a path of a manifest file")

	command.Flags().StringVarP(&options.Output, "output", "o", "", "path for the distribution archive (.tgz)")
	command.MarkFlagRequired("output")

	return command
}
