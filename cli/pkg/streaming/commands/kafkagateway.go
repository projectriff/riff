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

func NewKafkaGatewayCommand(ctx context.Context, c *cli.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "kafka-gateway",
		Short: "(experimental) kafka stream gateway",
		Long: strings.TrimSpace(`
The Kafka gateway encapsulates the address of a streaming gateway and a Kafka
provisioner instance.

The Kafka provisioner is responsible for creating topics in a Kafka cluster. The
streaming gateway coordinates and standardizes reads and writes to a Kafka
broker.
`),
		Aliases: []string{"kafka"},
	}

	cmd.AddCommand(NewKafkaGatewayListCommand(ctx, c))
	cmd.AddCommand(NewKafkaGatewayCreateCommand(ctx, c))
	cmd.AddCommand(NewKafkaGatewayDeleteCommand(ctx, c))
	cmd.AddCommand(NewKafkaGatewayStatusCommand(ctx, c))

	return cmd
}
