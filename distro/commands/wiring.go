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
	"github.com/projectriff/riff/pkg/fileutils"
	"github.com/projectriff/riff/pkg/resource"
	"os"

	"github.com/projectriff/riff/cmd/commands"

	"github.com/projectriff/riff/pkg/core"
	"github.com/projectriff/riff/pkg/docker"
	"github.com/spf13/cobra"
)

func DistroCreateAndWireRootCommand() *cobra.Command {

	var dockerClient docker.Docker
	var imageClient core.ImageClient

	rootCmd := &cobra.Command{
		Use:   "riff-distro",
		Short: "Commands for creating a riff distribution",
		Long: `riff is for functions.

riff is a CLI for functions on Knative.
See https://projectriff.io and https://github.com/knative/docs`,
		SilenceErrors:              true, // We'll print errors ourselves (after usage rather than before)
		DisableAutoGenTag:          true,
		SuggestionsMinimumDistance: 2,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			dockerClient = docker.RealDocker(os.Stdin, cmd.OutOrStdout(), cmd.OutOrStderr())
			imageClient = core.NewImageClient(dockerClient, fileutils.New(true), resource.ListImages)
			return nil
		},
	}

	image := Image()
	image.AddCommand(
		ImagePull(&imageClient),
		ImageList(&imageClient),
	)

	system := System()
	system.AddCommand(
		SystemDownload(&imageClient),
	)

	rootCmd.AddCommand(
		image,
		system,
		commands.Completion(rootCmd),
	)

	commands.Visit(rootCmd, func(c *cobra.Command) error {
		// Disable usage printing as soon as we enter RunE(), as errors that happen from then on
		// are not mis-usage error, but "regular" runtime errors
		exec := c.RunE
		if exec != nil {
			c.RunE = func(cmd *cobra.Command, args []string) error {
				c.SilenceUsage = true
				return exec(cmd, args)
			}
		}
		return nil
	})

	return rootCmd
}
