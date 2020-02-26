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

func NewStreamCommand(ctx context.Context, c *cli.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stream",
		Short: "(experimental) streams of messages",
		Long: strings.TrimSpace(`
A stream encapsulates an addressable message channel (typically a message 
broker's topic). It can be mapped to a function input or output stream.

Streams are managed by an associated streaming gateway and define a content 
type that its messages adhere to.
`),
		Aliases: []string{"streams"},
	}

	cmd.AddCommand(NewStreamListCommand(ctx, c))
	cmd.AddCommand(NewStreamCreateCommand(ctx, c))
	cmd.AddCommand(NewStreamDeleteCommand(ctx, c))
	cmd.AddCommand(NewStreamStatusCommand(ctx, c))

	return cmd
}
