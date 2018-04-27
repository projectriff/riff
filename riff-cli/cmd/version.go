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
	"strings"

	"github.com/projectriff/riff/riff-cli/global"
	invoker "github.com/projectriff/riff/riff-cli/pkg/invoker"
	"github.com/projectriff/riff/riff-cli/pkg/kubectl"
	"github.com/spf13/cobra"
)

func Version(w io.Writer, kubeCtl kubectl.KubeCtl) *cobra.Command {

	invokerOperations := invoker.Operations(kubeCtl)

	var versionCmd = &cobra.Command{
		Use:     "version",
		Short:   "Display the riff version",
		Long:    `Display the riff version`,
		Example: `  riff version`,

		Run: func(cmd *cobra.Command, args []string) {
			fmt.Fprintf(w, "riff CLI version: %v\n", global.CLI_VERSION)
			context, err := kubeCtl.Exec([]string{"config", "current-context"})
			if err != nil {
				context = "<unknown>"
			}
			fmt.Fprintf(w, "kubectl context: %v\n", strings.Trim(context, "\n"))
			fmt.Fprintln(w)
			components, err := kubeCtl.Exec([]string{
				"get", "deployments",
				"--all-namespaces",
				"-l", "app=riff",
				"--sort-by=metadata.labels.component",
				"-o=custom-columns=COMPONENT:.metadata.labels.component,IMAGE:.spec.template.spec.containers[0].image",
			})
			if err != nil {
				fmt.Fprint(w, "Unable to list components")
			} else {
				fmt.Fprintf(w, "%s", strings.Trim(components, "\n"))
			}
			fmt.Fprint(w, "\n\n")
			invokers, err := invokerOperations.Table()
			if err != nil {
				fmt.Fprint(w, "Unable to list invokers")
			} else {
				fmt.Fprintf(w, "%s", strings.Trim(invokers, "\n"))
			}
			fmt.Fprint(w, "\n")
		},
	}
	return versionCmd
}
