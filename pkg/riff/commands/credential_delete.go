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
	"github.com/projectriff/riff/pkg/riff"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type CredentialDeleteOptions struct {
	Namespace string
	Names     []string
	All       bool
}

func NewCredentialDeleteCommand(c *riff.Config) *cobra.Command {
	opt := &CredentialDeleteOptions{}

	cmd := &cobra.Command{
		Use: "delete",
		RunE: func(cmd *cobra.Command, args []string) error {
			client := c.Core().Secrets(opt.Namespace)

			if opt.All {
				return client.DeleteCollection(nil, metav1.ListOptions{
					// TODO get label from riff system
					LabelSelector: "projectriff.io/credential",
				})
			}

			for _, name := range opt.Names {
				// TODO check for the matching label before deleting
				if err := client.Delete(name, nil); err != nil {
					return err
				}
			}

			return nil
		},
	}

	riff.Args(cmd, riff.NamesArg(&opt.Names))

	riff.NamespaceFlag(cmd, c, &opt.Namespace)
	cmd.Flags().BoolVar(&opt.All, "all", false, "delete all secrets in the namespace")

	return cmd
}
