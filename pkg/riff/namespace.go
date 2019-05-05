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

package riff

import (
	"github.com/spf13/cobra"
)

func NamespaceFlag(cmd *cobra.Command, p *Params, namespace *string) {
	prior := cmd.PersistentPreRunE
	cmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		if *namespace == "" {
			*namespace = p.Client.DefaultNamespace
		}
		if prior != nil {
			return prior(cmd, args)
		}
		return nil
	}

	cmd.Flags().StringVarP(namespace, "namespace", "n", "", "the kubernetes namespace")
}