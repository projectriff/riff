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

// rootCmd represents the base command when called without any subcommands
func NewRiffCommand(p *riff.Params) *cobra.Command {
	var rootCmd = &cobra.Command{
		Use: "riff",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	rootCmd.PersistentFlags().StringVar(&p.ConfigFile, "config", "", "config file (default is $HOME/.riff.yaml)")
	rootCmd.PersistentFlags().StringVar(&p.KubeConfigFile, "kubeconfig", "", "kubectl config file (default is $HOME/.kube/config)")

	rootCmd.AddCommand(NewCredentialCommand(p))

	return rootCmd
}
