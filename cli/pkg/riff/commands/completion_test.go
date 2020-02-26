/*
 * Copyright 2019 The original author or authors
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

package commands_test

import (
	"strings"
	"testing"

	"github.com/projectriff/riff/cli/pkg/cli"
	"github.com/projectriff/riff/cli/pkg/riff/commands"
	rifftesting "github.com/projectriff/riff/cli/pkg/testing"
)

func TestCompletionOptions(t *testing.T) {
	table := rifftesting.OptionsTable{
		{
			Name:              "missing shell",
			Options:           &commands.CompletionOptions{},
			ExpectFieldErrors: cli.ErrMissingField(cli.ShellFlagName),
		},
		{
			Name: "invalid shell",
			Options: &commands.CompletionOptions{
				Shell: "zorglub",
			},
			ExpectFieldErrors: cli.ErrInvalidValue("zorglub", cli.ShellFlagName),
		},
		{
			Name: "valid shell bash",
			Options: &commands.CompletionOptions{
				Shell: "bash",
			},
			ShouldValidate: true,
		},
		{
			Name: "valid shell zsh",
			Options: &commands.CompletionOptions{
				Shell: "zsh",
			},
			ShouldValidate: true,
		},
	}

	table.Run(t)
}

func TestCompletionCommand(t *testing.T) {
	table := rifftesting.CommandTable{
		{
			Name: "default",
			Args: []string{},
			Verify: func(t *testing.T, output string, err error) {
				for _, str := range []string{
					"# bash completion",
				} {
					if !strings.Contains(output, str) {
						t.Errorf("expected completion output to contain %q\n", str)
					}
				}
			},
		},
		{
			Name: "bash",
			Args: []string{cli.ShellFlagName, "bash"},
			Verify: func(t *testing.T, output string, err error) {
				for _, str := range []string{
					"# bash completion",
				} {
					if !strings.Contains(output, str) {
						t.Errorf("expected completion output to contain %q\n", str)
					}
				}
			},
		},
		{
			Name: "zsh",
			Args: []string{cli.ShellFlagName, "zsh"},
			Verify: func(t *testing.T, output string, err error) {
				for _, str := range []string{
					"#compdef _completion completion",
				} {
					if !strings.Contains(output, str) {
						t.Errorf("expected completion output to contain %q\n", str)
					}
				}
			},
		},
	}

	table.Run(t, commands.NewCompletionCommand)
}
