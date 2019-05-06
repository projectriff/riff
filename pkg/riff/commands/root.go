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
	"github.com/projectriff/riff/pkg/env"
	"github.com/projectriff/riff/pkg/riff"
	"github.com/spf13/cobra"
)

func NewRootCommand(p *riff.Params) *cobra.Command {
	var cmd = &cobra.Command{
		Use: env.Cli.Name,
	}

	cmd.PersistentFlags().StringVar(&p.ConfigFile, "config", "", "config file (default is $HOME/.riff.yaml)")
	cmd.PersistentFlags().StringVar(&p.KubeConfigFile, "kubeconfig", "", "kubectl config file (default is $HOME/.kube/config)")

	cmd.AddCommand(NewCredentialCommand(p))
	cmd.AddCommand(NewApplicationCommand(p))
	cmd.AddCommand(NewFunctionCommand(p))
	cmd.AddCommand(NewRequestProcessorCommand(p))
	cmd.AddCommand(NewStreamCommand(p))
	cmd.AddCommand(NewStreamProcessorCommand(p))

	cmd.AddCommand(NewCompletionCommand(p))
	cmd.AddCommand(NewDocsCommand(p))
	cmd.AddCommand(NewVersionCommand(p))

	return cmd
}
