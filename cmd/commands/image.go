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
	"github.com/projectriff/riff/pkg/core"
	"github.com/spf13/cobra"
)

const (
	imageRelocateNumberOfArgs = iota
)

const (
	imageLoadNumberOfArgs = iota
)

const (
	imagePushNumberOfArgs = iota
)

func Image() *cobra.Command {
	return &cobra.Command{
		Use:   "image",
		Short: "Interact with docker images",
	}
}

func ImageRelocate(c *core.ImageClient) *cobra.Command {
	options := core.RelocateImagesOptions{}

	command := &cobra.Command{
		Use:   "relocate",
		Short: "Relocate docker image names to another registry",
		Long: "Relocate either a single kubernetes configuration file or a riff manifest, its kubernetes configuration files, " +
			"and an image manifest, so that image names refer to another (private or public) registry.\n" +

			"\nTo relocate a single kubernetes configuration file, use the `--file` flag to specify the path or URL of the file. Use " +
			"the `--output` flag to specify the path for the relocated file. If `--output` is an existing directory, the relocated " +
			"file will be placed in that directory. Otherwise the relocated file will be written to the path specified in `--output`.\n" +

			"\nTo relocate a manifest, use the `--manifest` flag to specify the path or URL of a manifest file which provides the paths or " +
			"URLs of the kubernetes configuration files for riff components. Use the `--output` flag to specify the path of a " +
			"directory to contain the relocated manifest, kubernetes configuration files, and image manifest. Any associated images " +
			"are copied to the output directory.\n" +

			"\nSpecify the registry hostname using the `--registry` flag, the user owning the images using the `--registry-user` flag, " +
			"and the images to be mapped using the `--images` flag. The `--images` flag contains the path of an " +
			"image manifest file, mapping image names to image ids, of the following form:\n" +
			`
    manifestVersion: 0.1
    images:
    ...
      "docker.io/istio/proxyv2:1.0.1": "sha256:7ae1462913665ac77389087f43d3d3dda86b4a0883b1cafcd105a4eeb648498f"
    ...
      "gcr.io/knative-releases/github.com/knative/serving/cmd/autoscaler@sha256:76222399addc02454db9837ea3ff54bae29849168586051a9d0180daa2c1a805": "sha256:76222399addc02454db9837ea3ff54bae29849168586051a9d0180daa2c1a805"
    ... 

`,
		Example: `  riff image relocate --manifest=path/to/manifest.yaml --registry=hostname --registry-user=username --images=path/to/image-manifest.yaml --output=path/to/output/dir
  riff image relocate --file=path/to/file --registry=hostname --registry-user=username --images=path/to/image-manifest.yaml --output=path/to/output`,
		Args: cobra.ExactArgs(imageRelocateNumberOfArgs),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if err := FlagsValidationConjunction(
				FlagsDependency(Set("manifest"), NoneOf("file")),
				FlagsDependency(Set("file"), NoneOf("manifest")),
			)(cmd); err != nil {
				return err
			}

			// validate --registry if it is set, otherwise allow flag omission to be diagnosed as such
			if cmd.Flags().Changed("registry") {
				if err := FlagValidName("registry")(cmd); err != nil {
					return err
				}
			}

			// prevent the default value of --manifest from conflicting with --file
			if cmd.Flags().Changed("file") && !cmd.Flags().Changed("manifest") {
				options.Manifest = ""
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			err := (*c).RelocateImages(options)
			if err != nil {
				return err
			}

			PrintSuccessfulCompletion(cmd)
			return nil
		},
	}

	command.Flags().StringVarP(&options.SingleFile, "file", "f", "", "path of a kubernetes configuration file")

	command.Flags().StringVarP(&options.Manifest, "manifest", "m", "manifest.yaml", "path of a riff manifest")

	command.Flags().StringVarP(&options.Registry, "registry", "r", "", "hostname for mapped images")
	command.MarkFlagRequired("registry")

	command.Flags().StringVarP(&options.RegistryUser, "registry-user", "u", "", "user name for mapped images")
	command.MarkFlagRequired("registry-user")

	command.Flags().StringVarP(&options.Images, "images", "i", "", "path of an image manifest")
	command.MarkFlagRequired("images")

	command.Flags().StringVarP(&options.Output, "output", "o", "", "path to contain the output file(s)")
	command.MarkFlagRequired("output")

	return command
}

func ImageLoad(c *core.ImageClient) *cobra.Command {
	options := core.LoadAndTagImagesOptions{}

	command := &cobra.Command{
		Use:   "load",
		Short: "Load and tag docker images",
		Long: "Load the images in an image manifest into a docker daemon and tag them.\n\n" +
			"For details of image manifests, see `riff image relocate -h`.\n\n" +
			"NOTE: This command requires the `docker` command line tool, as well as a docker daemon.\n\n" +
			"SEE ALSO: To load, tag, and push images to a registry, use `riff image push`.",
		Example: `  riff image load --images=riff-distro-xx/image-manifest.yaml`,
		Args:    cobra.ExactArgs(imageLoadNumberOfArgs),
		RunE: func(cmd *cobra.Command, args []string) error {
			err := (*c).LoadAndTagImages(options)
			if err != nil {
				return err
			}

			PrintSuccessfulCompletion(cmd)
			return nil
		},
	}
	command.Flags().StringVarP(&options.Images, "images", "i", "", "path of an image manifest")
	command.MarkFlagRequired("images")
	command.MarkFlagFilename("images", "yml", "yaml")

	return command
}

func ImagePush(c *core.ImageClient) *cobra.Command {
	options := core.PushImagesOptions{}

	command := &cobra.Command{
		Use:   "push",
		Short: "Push docker images to a registry",
		Long: "Load, tag, and push the images in an image manifest to a registry, for later consumption by `riff system install`.\n\n" +
			"For details of image manifests, see `riff image relocate -h`.\n\n" +
			"NOTE: This command requires the `docker` command line tool, as well as a docker daemon.\n\n" +
			"SEE ALSO: To load and tag images, but not push them, use `riff image load`.",
		Example: `  riff image push --images=riff-distro-xx/image-manifest.yaml`,
		Args:    cobra.ExactArgs(imagePushNumberOfArgs),
		RunE: func(cmd *cobra.Command, args []string) error {
			err := (*c).PushImages(options)
			if err != nil {
				return err
			}

			PrintSuccessfulCompletion(cmd)
			return nil
		},
	}
	command.Flags().StringVarP(&options.Images, "images", "i", "", "path of an image manifest")
	command.MarkFlagRequired("images")
	command.MarkFlagFilename("images", "yml", "yaml")

	return command
}
