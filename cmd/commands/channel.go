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
	"fmt"

	"github.com/projectriff/riff-cli/pkg/core"
	"github.com/spf13/cobra"
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

const (
	channelDeleteNameIndex = iota
	channelDeleteNumberOfArgs
)

var exactlyOneOfBusOrClusterBus = FlagsValidationConjunction(
	AtLeastOneOf("bus", "cluster-bus"),
	AtMostOneOf("bus", "cluster-bus"),
)

func ChannelCreate(fcTool *core.Client) *cobra.Command {
	options := core.CreateChannelOptions{}
	var write, force = false, false

	command := &cobra.Command{
		Use:   "create",
		Short: "create a new channel on a namespace or cluster bus",
		Args: ArgValidationConjunction(
			cobra.ExactArgs(channelCreateNumberOfArgs),
			AtPosition(channelCreateNameIndex, ValidName())),
		Example: `  riff channel create tweets --bus kafka --namespace steve-ns
  riff channel create orders --cluster-bus global-rabbit`,
		PreRunE: FlagsValidatorAsCobraRunE(exactlyOneOfBusOrClusterBus),
		RunE: func(cmd *cobra.Command, args []string) error {
			channelName := args[channelCreateNameIndex]
			options.Name = channelName

			c, err := (*fcTool).CreateChannel(options)
			if err != nil {
				return err
			}

			if write {
				marshaller, err := NewMarshaller(fmt.Sprintf("%s-channel.yaml", channelName), force)
				if err != nil {
					return err
				}
				if err = marshaller.Marshal(c); err != nil {
					return err
				}
			}

			return nil
		},
	}

	LabelArgs(command, "<channel-name>")

	command.Flags().StringVar(&options.Bus, "bus", "", busUsage)
	command.Flags().StringVar(&options.ClusterBus, "cluster-bus", "", clusterBusUsage)
	command.Flags().StringVarP(&options.Namespace, "namespace", "n", "", namespaceUsage)

	command.Flags().BoolVarP(&write, "write", "w", false, "whether to write yaml files for created resources")
	command.Flags().BoolVarP(&force, "force", "f", false, "force writing of files if they already exist")
	return command
}

func ChannelDelete(fcTool *core.Client) *cobra.Command {
	options := core.DeleteChannelOptions{}

	command := &cobra.Command{
		Use:   "delete",
		Short: "delete an existing channel",
		Args: ArgValidationConjunction(
			cobra.ExactArgs(channelDeleteNumberOfArgs),
			AtPosition(channelDeleteNameIndex, ValidName())),
		Example: `  riff channel delete tweets`,
		RunE: func(cmd *cobra.Command, args []string) error {
			options.Name = args[channelDeleteNameIndex]

			err := (*fcTool).DeleteChannel(options)
			return err
		},
	}

	LabelArgs(command, "<channel-name>")

	command.Flags().StringVarP(&options.Namespace, "namespace", "n", "", namespaceUsage)
	return command
}
