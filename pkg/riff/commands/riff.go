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

	"github.com/projectriff/riff/pkg/cli"
	"github.com/spf13/cobra"
)

func NewRiffCommand(ctx context.Context, c *cli.Config) *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "riff",
		Short: "riff is for functions",
		Long: strings.TrimSpace(`
<todo>
`),
		Args: cli.Args(),
	}

	cmd.AddCommand(NewCredentialCommand(ctx, c))
	cmd.AddCommand(NewApplicationCommand(ctx, c))
	cmd.AddCommand(NewFunctionCommand(ctx, c))
	cmd.AddCommand(NewHandlerCommand(ctx, c))
	cmd.AddCommand(NewStreamCommand(ctx, c))
	cmd.AddCommand(NewProcessorCommand(ctx, c))

	return cmd
}
