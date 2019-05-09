/*
 * Copyright 2019 the original author or authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      https://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"fmt"
	"os"
	"strings"

	// load credential helpers
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"

	"github.com/projectriff/riff/pkg/cli"
	"github.com/projectriff/riff/pkg/riff/commands"
)

func main() {
	cmd := commands.NewRootCommand(cli.Initialize())

	cmd.SilenceErrors = true
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(cmd.OutOrStderr(), "Error executing command:")
		// errors can be multiple lines, indent each line
		for _, line := range strings.Split(err.Error(), "\n") {
			fmt.Fprintf(cmd.OutOrStderr(), "  %s\n", line)
		}
		os.Exit(1)
	}
}
