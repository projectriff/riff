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
		Long: `Relocate either a single kubernetes configuration file or a riff manifest and its kubernetes
configuration files so that image names refer to another (private or public) registry.

To relocate a single kubernetes configuration file, use the '--file' flag to specify the path or URL of the file. Use
the '--output' flag to specify the path for the relocated file. If '--output' is an existing directory, the relocated
file will be placed in that directory. Otherwise the relocated file will be written to the path specified in '--output'.

To relocate a manifest, use the '--manifest' flag to specify the path of a manifest file which provides the paths or
URLs of the kubernetes configuration files for riff components. Use the '--output' flag to specify the path of a
directory to contain the relocated manifest and kubernetes configuration files.

Specify the registry hostname using the '--registry' flag, the user owning the images using the '--registry-user' flag,
and a complete list of the images to be mapped using the '--images' flag. The '--images' flag contains the path of an
image manifest file with contents of the following form:

    manifestVersion: 0.1
    images:
    ...
    - docker.io/istio/proxyv2:1.0.1
    ...
    - gcr.io/knative-releases/github.com/knative/serving/cmd/autoscaler@sha256:76222399addc02454db9837ea3ff54bae29849168586051a9d0180daa2c1a805
    ...
    
`,
		Example: `  riff image relocate --manifest=/path/to/manifest --registry=hostname --user=username --images=/path/to/image/manifest --output=/path/to/output/dir
  riff image relocate --file=/path/to/file --registry=hostname --user=username --images=/path/to/image/manifest --output=/path/to/output`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			// FIXME: these flags should not apply to this command: https://github.com/projectriff/riff/issues/743
			if cmd.Flags().Changed("kubeconfig") {
				return errors.New("the 'kubeconfig' flag is not supported by the 'image relocate' command")
			}
			m, _ := cmd.Flags().GetString("master")
			if len(m) > 0 {
				return errors.New("the 'master' flag is not supported by the 'image relocate' command")
			}

			return FlagsValidatorAsCobraRunE(ExactlyOneOf("file", "manifest"))(cmd, args)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			err := (*c).RelocateImages(options)
			if err != nil {
				return err
			}

			printSuccessfulCompletion(cmd)
			return nil
		},
	}

	command.Flags().StringVarP(&options.SingleFile, "file", "f", "", "path of a kubernetes configuration file")

	command.Flags().StringVarP(&options.Manifest, "manifest", "m", "manifest.yaml", "path of a riff manifest")

	command.Flags().StringVarP(&options.Registry, "registry", "r", "", "hostname for mapped images")
	command.MarkFlagRequired("registry")

	command.Flags().StringVarP(&options.RegistryUser, "registry-user", "u", "", "user name for mapped images")
	command.MarkFlagRequired("registry-user")

	command.Flags().StringVarP(&options.Images, "images", "i", "", "path of an image manifest of image names to be mapped")
	command.MarkFlagRequired("images")

	command.Flags().StringVarP(&options.Output, "output", "o", "", "path to contain the output file(s)")
	command.MarkFlagRequired("output")

	return command
}
