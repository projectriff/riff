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

package commands

import (
	"context"
	"strings"

	"github.com/projectriff/cli/pkg/cli"
	"github.com/spf13/cobra"
)

func NewApplicationCommand(ctx context.Context, c *cli.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "application",
		Short: "applications built from source using application buildpacks",
		Long: strings.TrimSpace(`
Applications are a mechanism to convert web application source code into
container images that can be invoked over HTTP. Cloud Native Buildpacks are
provided to detect the language, provide a language runtime, install build and
runtime dependencies, compile the application, and packaging everything as a
container.

The application resource is only responsible for converting source code into a
container. The application container image may then be deployed on the core or
knative runtime.
`),
		Aliases: []string{"applications", "app", "apps"},
	}

	cmd.AddCommand(NewApplicationListCommand(ctx, c))
	cmd.AddCommand(NewApplicationCreateCommand(ctx, c))
	cmd.AddCommand(NewApplicationDeleteCommand(ctx, c))
	cmd.AddCommand(NewApplicationStatusCommand(ctx, c))
	cmd.AddCommand(NewApplicationTailCommand(ctx, c))

	return cmd
}
