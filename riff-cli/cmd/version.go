/*
 * Copyright 2018 the original author or authors.
 *
 *   Licensed under the Apache License, Version 2.0 (the "License");
 *   you may not use this file except in compliance with the License.
 *   You may obtain a copy of the License at
 *
 *        http://www.apache.org/licenses/LICENSE-2.0
 *
 *   Unless required by applicable law or agreed to in writing, software
 *   distributed under the License is distributed on an "AS IS" BASIS,
 *   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *   See the License for the specific language governing permissions and
 *   limitations under the License.
 */

package cmd

import (
	"fmt"
	"io"

	"github.com/projectriff/riff/riff-cli/global"
	"github.com/projectriff/riff/riff-cli/pkg/kubectl"
	"github.com/spf13/cobra"
)

func Version(w io.Writer, kubeCtl kubectl.KubeCtl) *cobra.Command {
	// versionCmd represents the version command
	var versionCmd = &cobra.Command{
		Use:     "version",
		Short:   "Display the riff version",
		Long:    `Display the riff version`,
		Example: `  riff version`,

		Run: func(cmd *cobra.Command, args []string) {
			fmt.Fprintf(w, "riff CLI version: %v\n", global.CLI_VERSION)
			fmt.Fprintln(w)
			listing, err := kubeCtl.Exec([]string{
				"get", "deployments",
				"--all-namespaces",
				"-l", "app=riff",
				"-o=custom-columns=COMPONENT:.metadata.labels.component,IMAGE:.spec.template.spec.containers[0].image",
			})
			if err != nil {
				fmt.Fprint(w, "Unable to list component versions")
			} else {
				fmt.Fprintf(w, "%s", listing)
			}
		},
	}
	return versionCmd
}
