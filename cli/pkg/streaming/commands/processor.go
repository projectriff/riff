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

func NewProcessorCommand(ctx context.Context, c *cli.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "processor",
		Short: "(experimental) processors apply functions to messages on streams",
		Long: strings.TrimSpace(`
Processors coordinate reading from input streams and writing to output streams
with a function or container.

Function-based processors continuously watch for the latest built image and will
deploy new images. If the underlying build resource is deleted, the processor
will continue to run, but will no longer self update. Container-based processors
must be manually updated to trigger the rollout of an updated image.
`),
		Aliases: []string{"processors"},
	}

	cmd.AddCommand(NewProcessorListCommand(ctx, c))
	cmd.AddCommand(NewProcessorCreateCommand(ctx, c))
	cmd.AddCommand(NewProcessorDeleteCommand(ctx, c))
	cmd.AddCommand(NewProcessorStatusCommand(ctx, c))
	cmd.AddCommand(NewProcessorTailCommand(ctx, c))

	return cmd
}
