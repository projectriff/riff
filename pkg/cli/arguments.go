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
	"fmt"

	"github.com/spf13/cobra"
)

type Arg struct {
	Name  string
	Arity int
	Set   func(args []string) error
}

func Args(argDefs ...Arg) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		offset := 0

		for _, argDef := range argDefs {
			arity := argDef.Arity
			if arity == -1 {
				// consume all remaining args
				arity = len(args) - offset
			}
			if len(args) < arity {
				// TODO create a better message saying what is missing
				return fmt.Errorf("missing required argument(s)")
			}

			if err := argDef.Set(args[offset : offset+arity]); err != nil {
				return err
			}

			offset += arity
		}

		// no additional args
		return cobra.NoArgs(cmd, args[offset:])
	}
}

func NameArg(name *string) Arg {
	return Arg{
		Name:  "name",
		Arity: 1,
		Set: func(args []string) error {
			*name = args[0]
			return nil
		},
	}
}

func NamesArg(names *[]string) Arg {
	return Arg{
		Name:  "name(s)",
		Arity: -1,
		Set: func(args []string) error {
			*names = args
			return nil
		},
	}
}
