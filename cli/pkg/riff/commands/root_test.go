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
	"strings"
	"testing"

	"github.com/projectriff/cli/pkg/riff/commands"
	rifftesting "github.com/projectriff/cli/pkg/testing"
)

func TestRootCommand(t *testing.T) {
	table := rifftesting.CommandTable{
		{
			Name:        "empty",
			Args:        []string{},
			ShouldError: true,
		},
		{
			Name: "help",
			Args: []string{"--help"},
			Verify: func(t *testing.T, output string, err error) {
				if !strings.Contains(output, "riff [command]") {
					t.Errorf("expected help to contain command without args")
				}
			},
		},
		{
			Name: "help subcommand with args",
			Args: []string{"function", "create", "--help"},
			Verify: func(t *testing.T, output string, err error) {
				if !strings.Contains(output, "riff function create <name> [flags]") {
					t.Errorf("expected help to contain command with args")
				}
			},
		},
	}

	table.Run(t, commands.NewRootCommand)
}
