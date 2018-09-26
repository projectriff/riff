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
	"errors"
	"github.com/projectriff/riff/cmd/commands"
	"github.com/projectriff/riff/pkg/core"
	"github.com/spf13/cobra"
)

func System() *cobra.Command {
	return &cobra.Command{
		Use:   "system",
		Short: "Interact with riff systems",
	}
}

func SystemDownload(c *core.ImageClient) *cobra.Command {
	options := core.DownloadSystemOptions{}

	command := &cobra.Command{
		Use:   "download",
		Short: "Download a riff system.",
		Long: "Download the kubernetes configuration files for a given riff manifest.\n\n" +
			"Use the `--output` flag to specify the path of a " +
			"directory to contain the resultant kubernetes configuration files and rewritten riff manifest." +
			"The riff manifest is rewritten to refer to the downloaded configuration files.\n",
		Example: `  riff system download --manifest=path/to/manifest.yaml --output=path/to/output/dir`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				return errors.New("the `system download` command does not support positional arguments")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			err := (*c).DownloadSystem(options)
			if err != nil {
				return err
			}

			commands.PrintSuccessfulCompletion(cmd)
			return nil
		},
	}

	command.Flags().StringVarP(&options.Manifest, "manifest", "m", "stable", "manifest for the download; can be a named manifest (stable or latest) or a path of a manifest file")

	command.Flags().StringVarP(&options.Output, "output", "o", "", "path to contain the output file(s)")
	command.MarkFlagRequired("output")

	return command
}
