/*
 * Copyright 2018 The original author or authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package commands

import (
	"github.com/pivotal-cf-experimental/riff-cli/pkg/tool"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

func Channel() *cobra.Command {
	return &cobra.Command{
		Use:   "channel",
		Short: "interact with channel related resources",
	}
}

const (
	channelCreateNameIndex = iota
	channelCreateNumberOfArgs
)

var exactlyOneOfBusOrClusterBus = FlagsValidationConjunction(
	AtLeastOneOf("bus", "cluster-bus"),
	AtMostOneOf("bus", "cluster-bus"),
)

func ChannelCreate(fcTool *tool.Client) *cobra.Command {
	options := tool.CreateChannelOptions{}

	command := &cobra.Command{
		Use:   "create",
		Short: "create a new channel on a namespace or cluster bus",
		Args:  ArgValidationConjunction(cobra.ExactArgs(channelCreateNumberOfArgs), AtPosition(channelCreateNameIndex, ValidName())),
		Example: `  riff channel create tweets --bus kafka --namespace steve-ns
  riff channel create orders --cluster-bus global-rabbit`,
		PreRunE: FlagsValidatorAsCobraRunE(exactlyOneOfBusOrClusterBus),
		RunE: func(cmd *cobra.Command, args []string) error {
			options.Name = args[channelCreateNameIndex]

			c, err := (*fcTool).CreateChannel(options)

			e := yaml.NewEncoder(cmd.OutOrStdout())
			err = e.Encode(c)

			return err
		},
	}

	command.Flags().StringVar(&options.Bus, "bus", "", busUsage)
	command.Flags().StringVar(&options.ClusterBus, "cluster-bus", "", clusterBusUsage)
	command.Flags().StringVarP(&options.Namespace, "namespace", "n", "", namespaceUsage)
	return command
}
