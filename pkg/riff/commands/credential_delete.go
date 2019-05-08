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
)

type CredentialDeleteOptions struct {
	Namespace string
	Name      string
}

func NewCredentialDeleteCommand(p *riff.Params) *cobra.Command {
	opt := &CredentialDeleteOptions{}

	cmd := &cobra.Command{
		Use: "delete",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			// TODO validate arg
			opt.Name = args[0]
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return p.Core().Secrets(opt.Namespace).Delete(opt.Name, nil)
		},
	}

	riff.NamespaceFlag(cmd, p, &opt.Namespace)

	return cmd
}
