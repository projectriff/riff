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

	buildcommands "github.com/projectriff/riff/cli/pkg/build/commands"
	"github.com/projectriff/riff/cli/pkg/cli"
	"github.com/spf13/cobra"
)

func NewRiffCommand(ctx context.Context, c *cli.Config) *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "riff",
		Short: "riff is for functions",
		Long: strings.TrimSpace(`
The ` + c.Name + ` CLI combines with the projectriff system CRDs to build, run and wire
workloads (functions, applications and containers). The CRDs provide the riff
API of which this CLI is a client.

Before running ` + c.Name + `, please install the projectriff system and its dependencies.
See https://projectriff.io/docs/getting-started/

The application, function and container commands define build plans and the
credential commands to authenticate builds to container registries.

Runtimes provide ways to execute the workloads. Different runtimes provide
alternate execution models and capabilities.
`),
	}

	cmd.AddCommand(buildcommands.NewCredentialCommand(ctx, c))
	cmd.AddCommand(buildcommands.NewApplicationCommand(ctx, c))
	cmd.AddCommand(buildcommands.NewContainerCommand(ctx, c))
	cmd.AddCommand(buildcommands.NewFunctionCommand(ctx, c))

	return cmd
}
