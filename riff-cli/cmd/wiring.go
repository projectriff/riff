/*
 * Copyright 2017 the original author or authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package cmd

import (
	"fmt"
	"os"

	"github.com/projectriff/riff/riff-cli/pkg/docker"
	invoker "github.com/projectriff/riff/riff-cli/pkg/invoker"
	"github.com/projectriff/riff/riff-cli/pkg/kubectl"
	"github.com/projectriff/riff/riff-cli/pkg/minikube"
	"github.com/spf13/cobra"
)

// CreateAndWireRootCommand creates all riff commands and sub commands, as well as the top-level 'root' command,
// wires them together and returns the root command, ready to execute.
func CreateAndWireRootCommand(realDocker docker.Docker, dryRunDocker docker.Docker,
	realKubeCtl kubectl.KubeCtl, dryRunKubeCtl kubectl.KubeCtl,
	minik minikube.Minikube) (*cobra.Command, error) {

	invokerOperations := invoker.Operations(realKubeCtl)
	invokers, err := invokerOperations.List()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Unable to load invokers via kubectl")
	}

	rootCmd := Root()

	initCmd, initOptions := Init(invokers)
	initInvokerCmds, err := InitInvokers(invokers, initOptions)
	if err != nil {
		return nil, err
	}
	initCmd.AddCommand(initInvokerCmds...)

	buildCmd, _ := Build(realDocker, dryRunDocker)

	applyCmd, _ := Apply(realKubeCtl, dryRunKubeCtl)

	createCmd := Create(initCmd, buildCmd, applyCmd)
	createInvokerCmds := CreateInvokers(invokers, initInvokerCmds, buildCmd, applyCmd)
	createCmd.AddCommand(createInvokerCmds...)

	deleteCmd, _ := Delete(realKubeCtl, dryRunKubeCtl)

	invokersCmd := Invokers()
	invokersApplyCmd, _ := InvokersApply(realKubeCtl)
	invokersListCmd := InvokersList(realKubeCtl)
	invokersDeleteCmd, _ := InvokersDelete(realKubeCtl)
	invokersCmd.AddCommand(invokersApplyCmd, invokersListCmd, invokersDeleteCmd)

	rootCmd.AddCommand(
		applyCmd,
		buildCmd,
		createCmd,
		deleteCmd,
		initCmd,
		List(realKubeCtl),
		Logs(realKubeCtl),
		Publish(realKubeCtl, minik),
		Update(buildCmd, applyCmd),
		invokersCmd,
		Version(os.Stdout, realKubeCtl),
	)

	rootCmd.AddCommand(
		Completion(rootCmd),
		Docs(rootCmd),
	)

	return rootCmd, nil
}
