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
	"errors"

	"github.com/projectriff/riff/pkg/core"
	"github.com/spf13/cobra"
)

func Image() *cobra.Command {
	return &cobra.Command{
		Use:   "image",
		Short: "Interact with docker images",
	}
}

func ImageRelocate(c *core.Client) *cobra.Command {
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
			"image manifest file with contents of the following form:\n" +
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
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				return errors.New("the `image relocate` command does not support positional arguments")
			}

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

	command.Flags().BoolVar(&options.Flatten, "flatten", false, "flatten image names (for registries that do not support hierarchical names)")

	command.Flags().StringVarP(&options.Images, "images", "i", "", "path of an image manifest of image names to be mapped")
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
		Long: "Load the set of images identified by the provided image manifest into a docker daemon.\n\n" +
			"NOTE: This command requires the `docker` command line tool, as well as a (local) docker daemon.\n\n" +
			"SEE ALSO: To load, tag, and push images, use `riff image push`.",
		Example: `  riff image load --images=riff-distro-xx/image-manifest.yaml`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			// FIXME: these flags should not apply to this command: https://github.com/projectriff/riff/issues/743
			if cmd.Flags().Changed("kubeconfig") {
				return errors.New("the 'kubeconfig' flag is not supported by the 'image load' command")
			}
			m, _ := cmd.Flags().GetString("master")
			if len(m) > 0 {
				return errors.New("the 'master' flag is not supported by the 'image load' command")
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			err := (*c).LoadAndTagImages(options)
			if err != nil {
				return err
			}

			PrintSuccessfulCompletion(cmd)
			return nil
		},
	}
	command.Flags().StringVarP(&options.Images, "images", "i", "", "path of an image manifest of image names to be loaded")
	command.MarkFlagRequired("images")
	command.MarkFlagFilename("images", "yml", "yaml")

	return command
}

func ImagePush(c *core.ImageClient) *cobra.Command {
	options := core.PushImagesOptions{}

	command := &cobra.Command{
		Use:   "push",
		Short: "Push (relocated) docker image names to an image registry",
		Long: "Push the set of images identified by the provided image manifest into a remote registry, for later consumption by `riff system install`.\n\n" +
			"NOTE: This command requires the `docker` command line tool, as well as a (local) docker daemon and will load and tag the images using that daemon.\n\n" +
			"SEE ALSO: To load and tag images, but not push them, use `riff image load`.",
		Example: `  riff image push --images=riff-distro-xx/image-manifest.yaml`,
		RunE: func(cmd *cobra.Command, args []string) error {
			err := (*c).PushImages(options)
			if err != nil {
				return err
			}

			PrintSuccessfulCompletion(cmd)
			return nil
		},
	}
	command.Flags().StringVarP(&options.Images, "images", "i", "", "path of an image manifest of image names to be pushed")
	command.MarkFlagRequired("images")
	command.MarkFlagFilename("images", "yml", "yaml")

	return command
}
