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

package commands_test

import (
	"context"
	"github.com/projectriff/riff/pkg/cli"
	"strings"
	"testing"

	"github.com/projectriff/riff/pkg/riff/commands"
	rifftesting "github.com/projectriff/riff/pkg/testing"
	"github.com/spf13/cobra"
)

func TestRiffCommand(t *testing.T) {
	table := rifftesting.CommandTable{
		{
			Name: "empty",
			Args: []string{},
		},
	}

	table.Run(t, commands.NewRiffCommand)
}

func TestRiffSubCommands(t *testing.T) {
	riffCommand := commands.NewRiffCommand(context.Background(), cli.NewDefaultConfig())
	commands := riffCommand.Commands()

	commandsUses := []string{
		// "application",
		// "credential",
		"doctor",
		// "function",
		// "handler",
		// "processor",
		// "stream",
	}

	for _, commandUse := range commandsUses {
		var commandFound = findCommandByUse(commandUse, commands)
		for _, command := range commands {
			if strings.ToLower(command.Use) == strings.ToLower(commandUse) {
				commandToFind = command
			}
		}

		if commandFound == nil {
			t.Fatalf("No %s command in riff command list", commandUse)
		}

		if commandFound.Short == "" {
			t.Fatalf("%s command has no Short description", commandUse)
		}

		if commandFound.Long == "" {
			t.Fatalf("%s command has no Long description", commandUse)
		}

		if commandFound.Example == "" {
			t.Fatalf("%s command has no Example", commandUse)
		}
	}
}
