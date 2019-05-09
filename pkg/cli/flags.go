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

package cli

import (
	"github.com/spf13/cobra"
)

func AllNamespacesFlag(cmd *cobra.Command, c *Config, namespace *string, allNamespaces *bool) {
	prior := cmd.PreRunE
	cmd.PreRunE = func(cmd *cobra.Command, args []string) error {
		if *allNamespaces == true {
			*namespace = ""
		}
		if prior != nil {
			if err := prior(cmd, args); err != nil {
				return err
			}
		}
		return nil
	}

	NamespaceFlag(cmd, c, namespace)
	cmd.Flags().BoolVar(allNamespaces, "all-namespaces", false, "list the requested object(s) across all namespaces")
}

func NamespaceFlag(cmd *cobra.Command, c *Config, namespace *string) {
	prior := cmd.PreRunE
	cmd.PreRunE = func(cmd *cobra.Command, args []string) error {
		if *namespace == "" {
			*namespace = c.DefaultNamespace()
		}
		if prior != nil {
			if err := prior(cmd, args); err != nil {
				return err
			}
		}
		return nil
	}

	cmd.Flags().StringVarP(namespace, "namespace", "n", "", "the kubernetes namespace")
}
