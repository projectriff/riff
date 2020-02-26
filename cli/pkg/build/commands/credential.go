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

func NewCredentialCommand(ctx context.Context, c *cli.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "credential",
		Short: "credentials for container registries",
		Long: strings.TrimSpace(`
Credentials allow builds to push images to authenticated registries. If the
registry allows unauthenticated image pushes, credentials are not required
(while useful for local development environments, this is not recommended).

Credentials are defined by a hostname, username and password. These values are
specified explicitly or via shortcuts for Docker Hub and Google Container
Registry (GCR).

The credentials are saved as Kubernetes secrets and exposed to build pods.

To manage credentials, read and write access to Secrets is required for the
namespace. To manage the default image prefix, read and write access to the
'riff-build' ConfigMap is required for the namespace.
`),
		Aliases: []string{"credentials", "cred", "creds"},
	}

	cmd.AddCommand(NewCredentialListCommand(ctx, c))
	cmd.AddCommand(NewCredentialApplyCommand(ctx, c))
	cmd.AddCommand(NewCredentialDeleteCommand(ctx, c))

	return cmd
}
