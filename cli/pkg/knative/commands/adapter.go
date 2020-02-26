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

func NewAdapterCommand(ctx context.Context, c *cli.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "adapter",
		Short: "adapters push built images to Knative",
		Long: strings.TrimSpace(`
The Knative runtime adapter updates a Knative Service or Configuration with the
latest image from a riff build. As the build produces new images, they will be
rolled out automatically to the target Knative resource.

No new Knative resources are created directly by the adapter, it only updates
the image for an existing resource.
`),
		Aliases: []string{"adapters"},
	}

	cmd.AddCommand(NewAdapterListCommand(ctx, c))
	cmd.AddCommand(NewAdapterCreateCommand(ctx, c))
	cmd.AddCommand(NewAdapterDeleteCommand(ctx, c))
	cmd.AddCommand(NewAdapterStatusCommand(ctx, c))

	return cmd
}
