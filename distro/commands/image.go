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

const (
	imagePullNumberOfArgs = iota
)

const (
	imageListNumberOfArgs = iota
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
		Example: `  riff-distro image pull --images=riff-distro-xx/image-manifest.yaml`,
		Args: cobra.ExactArgs(imagePullNumberOfArgs),
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

func ImageList(c *core.ImageClient) *cobra.Command {
	options := core.ListImagesOptions{}

	/*
			searches an input manifest and associated k8s files for image names and creates an image manifest listing the images.

		It does not guarantee to find all images referenced by the k8s files and so the resultant list of images needs to be validated by the user, e.g. by manual inspection or testing.
	*/
	command := &cobra.Command{
		Use:   "list",
		Short: "List some or all of the images for a riff manifest",
		Long: "Search a riff manifest and associated kubernetes configuration files for image names and create an image manifest listing the images.\n\n" +
			"It does not guarantee to find all referenced images and so the resultant image manifest needs to be validated, for example by manual inspection or testing.\n\n" +
			"NOTE: This command requires the `docker` command line tool to check the images.",
		Example: `  riff-distro image list --manifest=path/to/manifest.yaml --images=path/for/image-manifest.yaml`,
		Args: cobra.ExactArgs(imageListNumberOfArgs),
		RunE: func(cmd *cobra.Command, args []string) error {
			err := (*c).ListImages(options)
			if err != nil {
				return err
			}

			commands.PrintSuccessfulCompletion(cmd)
			return nil
		},
	}
	command.Flags().StringVarP(&options.Manifest, "manifest", "m", "stable", "manifest to be searched; can be a named manifest (stable or latest) or a path of a manifest file")
	command.MarkFlagFilename("manifest", "yml", "yaml")

	command.Flags().StringVarP(&options.Images, "images", "i", "", "path of the image manifest to be created; defaults to 'image-manifest.yaml' relative to the manifest")
	command.MarkFlagFilename("images", "yml", "yaml")

	command.Flags().BoolVarP(&options.NoCheck, "no-check", "", false, "skips checking the images, thus not omitting the ones unknown to docker")
	command.Flag("no-check").NoOptDefVal = "true"
	command.Flags().BoolVarP(&options.Force, "force", "", false, "overwrite the image manifest if it already exists")

	return command
}
