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
	"fmt"

	"github.com/projectriff/riff/pkg/riff"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type CredentialListOptions struct {
	Namespace     string
	AllNamespaces bool
}

func NewCredentialListCommand(c *riff.Config) *cobra.Command {
	opt := &CredentialListOptions{}

	cmd := &cobra.Command{
		Use: "list",
		RunE: func(cmd *cobra.Command, args []string) error {
			secrets, err := c.Core().Secrets(opt.Namespace).List(metav1.ListOptions{
				// TODO get label from riff system
				LabelSelector: "projectriff.io/credential",
			})
			if err != nil {
				return err
			}

			if len(secrets.Items) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No credentials found.")
			}
			for _, secret := range secrets.Items {
				// TODO pick a generic table formatter
				fmt.Fprintln(cmd.OutOrStdout(), secret.Name)
			}

			return nil
		},
	}

	riff.AllNamespacesFlag(cmd, c, &opt.Namespace, &opt.AllNamespaces)

	return cmd
}
