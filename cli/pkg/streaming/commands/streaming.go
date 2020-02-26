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

	"github.com/projectriff/riff/cli/pkg/cli"
	"github.com/spf13/cobra"
)

func NewStreamingCommand(ctx context.Context, c *cli.Config) *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "streaming",
		Short: "(experimental) streaming runtime for " + c.Name + " functions",
		Long: strings.TrimSpace(`
The streaming runtime uses ` + c.Name + ` functions, processor and stream custom resources
to deploy streaming workloads. 

Functions can accept several input and/or output streams.
`),
	}

	cmd.AddCommand(NewStreamCommand(ctx, c))
	cmd.AddCommand(NewProcessorCommand(ctx, c))
	cmd.AddCommand(NewGatewayCommand(ctx, c))
	cmd.AddCommand(NewInMemoryGatewayCommand(ctx, c))
	cmd.AddCommand(NewKafkaGatewayCommand(ctx, c))
	cmd.AddCommand(NewPulsarGatewayCommand(ctx, c))

	return cmd
}
