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
	"bufio"
	"context"
	"io/ioutil"
	"syscall"

	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh/terminal"
)

func Sequence(items ...func(cmd *cobra.Command, args []string) error) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		for i := range items {
			if err := items[i](cmd, args); err != nil {
				return err
			}
		}
		return nil
	}
}

func Visit(cmd *cobra.Command, f func(c *cobra.Command) error) error {
	err := f(cmd)
	if err != nil {
		return err
	}
	for _, c := range cmd.Commands() {
		err := Visit(c, f)
		if err != nil {
			return err
		}
	}
	return nil
}

func ReadStdin(c *Config, value *[]byte, prompt string) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		if terminal.IsTerminal(int(syscall.Stdin)) {
			c.Printf("%s: ", prompt)
			res, err := terminal.ReadPassword(int(syscall.Stdin))
			c.Printf("\n")
			if err != nil {
				return err
			}
			*value = res
		} else {
			reader := bufio.NewReader(c.Stdin)
			res, err := ioutil.ReadAll(reader)
			if err != nil {
				return err
			}
			*value = res
		}

		return nil
	}
}

type commandKey struct{}

func WithCommand(ctx context.Context, cmd *cobra.Command) context.Context {
	return context.WithValue(ctx, commandKey{}, cmd)
}

func CommandFromContext(ctx context.Context) *cobra.Command {
	if cmd, ok := ctx.Value(commandKey{}).(*cobra.Command); ok {
		return cmd
	}
	return nil
}
