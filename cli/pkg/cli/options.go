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
	"context"

	"github.com/spf13/cobra"
)

// ValidateOptions bridges a cobra RunE function to the Validatable interface.  All flags and
// arguments must already be bound, with explicit or default values, to the options struct being
// validated. This function is typically used to define the PreRunE phase of a command.
//
// ```
// cmd := &cobra.Command{
// 	   ...
// 	   PreRunE: cli.ValidateOptions(opts),
// }
// ```
func ValidateOptions(ctx context.Context, opts Validatable) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		ctx := WithCommand(ctx, cmd)
		if err := opts.Validate(ctx); len(err) != 0 {
			return err.ToAggregate()
		}
		cmd.SilenceUsage = true
		return nil
	}
}

type Executable interface {
	Exec(ctx context.Context, c *Config) error
}

// ExecOptions bridges a cobra RunE function to the Executable interface.  All flags and
// arguments must already be bound, with explicit or default values, to the options struct being
// validated. This function is typically used to define the RunE phase of a command.
//
// ```
// cmd := &cobra.Command{
// 	   ...
// 	   RunE: cli.ExecOptions(c, opts),
// }
// ```
func ExecOptions(ctx context.Context, c *Config, opts Executable) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		ctx := WithCommand(ctx, cmd)
		if o, ok := opts.(DryRunable); ok && o.IsDryRun() {
			// reserve Stdout for resources, redirect normal stdout to stderr
			ctx = withStdout(ctx, c.Stdout)
			c.Stdout = c.Stderr
		}
		return opts.Exec(ctx, c)
	}
}
