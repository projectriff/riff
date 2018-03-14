/*
 * Copyright 2018 the original author or authors.
 *
 *   Licensed under the Apache License, Version 2.0 (the "License");
 *   you may not use this file except in compliance with the License.
 *   You may obtain a copy of the License at
 *
 *        http://www.apache.org/licenses/LICENSE-2.0
 *
 *   Unless required by applicable law or agreed to in writing, software
 *   distributed under the License is distributed on an "AS IS" BASIS,
 *   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *   See the License for the specific language governing permissions and
 *   limitations under the License.
 */

package cmd

import (
	"fmt"

	projectriff_v1 "github.com/projectriff/riff/kubernetes-crds/pkg/apis/projectriff.io/v1"
	"github.com/projectriff/riff/riff-cli/cmd/utils"
	"github.com/spf13/cobra"
)

func Create(initCmd *cobra.Command, buildCmd *cobra.Command, applyCmd *cobra.Command) *cobra.Command {
	createChainCmd := utils.CommandChain(initCmd, buildCmd, applyCmd)
	createChainCmd.Use = "create"
	createChainCmd.Short = "Create a function"
	createChainCmd.Long = utils.CreateCmdLong()
	return createChainCmd
}

func CreateInvokers(invokers []projectriff_v1.Invoker, initInvokerCmds []*cobra.Command, buildCmd *cobra.Command, applyCmd *cobra.Command) []*cobra.Command {
	var createInvokerCmds []*cobra.Command

	for i, invoker := range invokers {
		invokerName := invoker.ObjectMeta.Name
		initInvokerCmd := initInvokerCmds[i]
		createInvokerCmd := utils.CommandChain(initInvokerCmd, buildCmd, applyCmd)
		createInvokerCmd.Use = invokerName
		createInvokerCmd.Short = fmt.Sprintf("Create a %s function", invokerName)
		createInvokerCmd.Long = utils.CreateInvokerCmdLong(invoker)

		createInvokerCmds = append(createInvokerCmds, createInvokerCmd)
	}

	return createInvokerCmds
}
