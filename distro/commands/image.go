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
	"github.com/spf13/cobra"
)

func Image() *cobra.Command {
	return &cobra.Command{
		Use:   "image",
		Short: "Interact with docker images",
	}
}

func ImagePull(c *core.ImageClient) *cobra.Command {
	options := core.PullImagesOptions{}

	command := &cobra.Command{
		Use:   "pull",
		Short: "Pull all docker images referenced in a distribution image-manifest and write them to disk",
		Long: "Pull the set of images identified by the provided image manifest from remote registries, in preparation of an offline distribution tarball.\n\n" +
			"NOTE: This command requires the `docker` command line tool, as well as a (local) docker daemon and will load and tag the images using that daemon.",
		Example: `  riff image pull --images=riff-distro-xx/image-manifest.yaml`,
		RunE: func(cmd *cobra.Command, args []string) error {
			err := (*c).PullImages(options)
			if err != nil {
				return err
			}

			commands.PrintSuccessfulCompletion(cmd)
			return nil
		},
	}
	command.Flags().StringVarP(&options.Images, "images", "i", "", "path of an image manifest of image names to be pulled")
	command.MarkFlagRequired("images")
	command.MarkFlagFilename("images", "yml", "yaml")

	command.Flags().StringVarP(&options.Output, "output", "o", "", "output `directory` for both the new manifest and images; defaults to rewriting the manifest in place with a sibling images/ directory")
	command.Flags().BoolVarP(&options.ContinueOnMismatch, "continue", "c", false, "whether to continue if an image doesn't have the same digest as stated in the image manifest; fail otherwise")
	return command
}
