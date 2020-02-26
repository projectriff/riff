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
	"strconv"

	"github.com/spf13/cobra"
)

const (
	NameArgumentName  = "name"
	NamesArgumentName = "name(s)"
)

var ErrIgnoreArg = fmt.Errorf("ignore argument")

type Arg struct {
	Name     string
	Arity    int
	Optional bool
	Set      func(cmd *cobra.Command, args []string, offset int) error
}

func Args(cmd *cobra.Command, argDefs ...Arg) {
	cmd.Args = func(cmd *cobra.Command, args []string) error {
		offset := 0

		for _, argDef := range argDefs {
			arity := argDef.Arity
			if arity == -1 {
				// consume all remaining args
				arity = len(args) - offset
			}
			if len(args)-offset < arity {
				if argDef.Optional {
					continue
				}
				// TODO create a better message saying what is missing
				return fmt.Errorf("missing required argument(s)")
			}

			if err := argDef.Set(cmd, args, offset); err != nil {
				if err == ErrIgnoreArg {
					continue
				}
				return err
			}

			offset += arity
		}

		// no additional args
		return cobra.NoArgs(cmd, args[offset:])
	}

	if cmd.Annotations == nil {
		cmd.Annotations = map[string]string{}
	}

	// add docs annoations
	cmd.Annotations["args.length"] = fmt.Sprintf("%d", len(argDefs))
	for i, argDef := range argDefs {
		cmd.Annotations[fmt.Sprintf("args[%d].name", i)] = argDef.Name
		cmd.Annotations[fmt.Sprintf("args[%d].arity", i)] = fmt.Sprintf("%d", argDef.Arity)
		cmd.Annotations[fmt.Sprintf("args[%d].optional", i)] = fmt.Sprintf("%v", argDef.Optional)
	}
}

func FormatArgs(cmd *cobra.Command) string {
	str := ""
	length, _ := strconv.Atoi(cmd.Annotations["args.length"])
	for i := 0; i < length; i++ {
		name := cmd.Annotations[fmt.Sprintf("args[%d].name", i)]
		if name == "" {
			continue
		}

		if cmd.Annotations[fmt.Sprintf("args[%d].optional", i)] == "true" {
			name = fmt.Sprintf("[%s]", name)
		} else {
			name = fmt.Sprintf("<%s>", name)
		}

		str += " " + name
	}

	return str
}

func NameArg(name *string) Arg {
	return Arg{
		Name:  NameArgumentName,
		Arity: 1,
		Set: func(cmd *cobra.Command, args []string, offset int) error {
			*name = args[offset]
			return nil
		},
	}
}

func NamesArg(names *[]string) Arg {
	return Arg{
		Name:  NamesArgumentName,
		Arity: -1,
		Set: func(cmd *cobra.Command, args []string, offset int) error {
			*names = args[offset:]
			return nil
		},
	}
}

func BareDoubleDashArgs(values *[]string) Arg {
	return Arg{
		Arity: -1,
		Set: func(cmd *cobra.Command, args []string, offset int) error {
			if cmd.ArgsLenAtDash() == -1 {
				return nil
			}
			*values = args[cmd.ArgsLenAtDash():]
			return nil
		},
	}
}
