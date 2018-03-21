/*
 * Copyright 2018 the original author or authors.
 *
 *   Licensed under the Apache License, Version 2.0 (the "License");
 *   you may not use this file except in compliance with the License.
 *   You may obtain a copy of the License at
 *
 *        http://www.apache.org/licenses/LICENSE-2.0
 *
 *   Unless required by applicable law or agreed to in writing, software
 *   distributed under the License is distributed on an "AS IS" BASIS,
 *   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *   See the License for the specific language governing permissions and
 *   limitations under the License.
 */

package utils

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"fmt"
)

// CommandChain returns a composite command that runs the provided commands one after the other.
// For each run kind, the variants that can error (ie runE vs run) are preferred if defined.
// postRun variations are run in reversed order
func CommandChain(commands ... *cobra.Command) *cobra.Command {

	argCache := make(map[*cobra.Command][]string)

	persistentPreRunE := func(cmd *cobra.Command, args []string) error {
		for _, command := range commands {
			command.ParseFlags(args)
			argCache[command] = command.Flags().Args()

			for p := command; p != nil; p = p.Parent() {
				if p.PersistentPreRunE != nil {
					if err := p.PersistentPreRunE(cmd, argCache[command]); err != nil {
						return err
					}
				} else if p.PersistentPreRun != nil {
					p.PersistentPreRun(cmd, argCache[command])
				}
			}
		}
		return nil
	}
	persistentPreRun := func(cmd *cobra.Command, args []string) {
		persistentPreRunE(cmd, args)
	}

	preRunE := func(cmd *cobra.Command, _ []string) error {
		for _, command := range commands {
			if command.PreRunE != nil {
				err := command.PreRunE(cmd, argCache[command])
				if err != nil {
					return err
				}
			} else if command.PreRun != nil {
				command.PreRun(cmd, argCache[command])
			}
		}
		return nil
	}
	preRun := func(cmd *cobra.Command, args []string) {
		preRunE(cmd, args)
	}

	runE := func(cmd *cobra.Command, _ []string) error {
		for _, command := range commands {
			if command.RunE != nil {
				err := command.RunE(cmd, argCache[command])
				if err != nil {
					return err
				}
			} else {
				command.Run(cmd, argCache[command])
			}
		}
		return nil
	}
	run := func(cmd *cobra.Command, args []string) {
		runE(cmd, args)
	}

	postRunE := func(cmd *cobra.Command, _ []string) error {
		for i := len(commands) - 1; i >= 0; i-- {
			command := commands[i]
			if command.PostRunE != nil {
				err := command.PostRunE(cmd, argCache[command])
				if err != nil {
					return err
				}
			} else if command.PostRun != nil {
				command.PostRun(cmd, argCache[command])
			}
		}
		return nil
	}
	postRun := func(cmd *cobra.Command, args []string) {
		postRunE(cmd, args)
	}

	persistentPostRunE := func(cmd *cobra.Command, _ []string) error {
		for i := len(commands) - 1; i >= 0; i-- {
			command := commands[i]
			for p := command; p != nil; p = p.Parent() {
				if p.PersistentPostRunE != nil {
					if err := p.PersistentPostRunE(cmd, argCache[command]); err != nil {
						return err
					}
				} else if p.PersistentPostRun != nil {
					p.PersistentPostRun(cmd, argCache[command])
				}
			}
		}
		return nil
	}
	persistentPostRun := func(cmd *cobra.Command, args []string) {
		persistentPostRunE(cmd, args)
	}

	var validators []cobra.PositionalArgs
	for _, command := range commands {
		if command.Args != nil {
			validators = append(validators, command.Args)
		}
	}

	var chain = &cobra.Command{
		Run:                run,
		RunE:               runE,
		PreRun:             preRun,
		PreRunE:            preRunE,
		PostRun:            postRun,
		PostRunE:           postRunE,
		PersistentPreRun:   persistentPreRun,
		PersistentPreRunE:  persistentPreRunE,
		PersistentPostRun:  persistentPostRun,
		PersistentPostRunE: persistentPostRunE,
		Args:               And(validators...),
	}

	// The flags for the chain command will look like the union of all the flags of delegate commands, with each
	// flag, if it is repeated, broadcasting its Set() call to each delegate flag.
	// So if `update = build + apply` and both 'build' and 'apply' have a --filepath flag, then setting that flag
	// ends up calling both build's and apply's flag.Set(), each writing to their own pointed value.
	// Duplicated flags are checked for meaning equality and the function panics if they differ
	for _, c := range commands {
		// This forces correct initialization and inheritance of c.Flags() (which c.Flags() documentation
		// advertises but actually doesn't do)
		c.LocalNonPersistentFlags()
		c.InheritedFlags()

		copyFlagsToChain(*c.PersistentFlags(), chain.PersistentFlags())
		copyFlagsToChain(*c.Flags(), chain.Flags())
	}
	return chain
}

func copyFlagsToChain(commandFlags pflag.FlagSet, chainFlags *pflag.FlagSet) {
	commandFlags.VisitAll(func(f *pflag.Flag) {
		flag := chainFlags.Lookup(f.Name)
		if flag == nil {
			chainFlags.AddFlag(newBroadcastFlag(f))
		} else {
			checkFlagConsistency(flag, f)
			flag.Value = append(flag.Value.(broadcastValue), f.Value)
		}
	})

}

func checkFlagConsistency(a *pflag.Flag, b *pflag.Flag) {
	if a.Usage != b.Usage ||
		a.Shorthand != b.Shorthand ||
		a.DefValue != b.DefValue ||
		a.NoOptDefVal != b.NoOptDefVal {
		panic(fmt.Sprintf("Trying to chain together methods with different flags with the same name:\n%v\n%v", a, b))
	}
}

func newBroadcastFlag(f *pflag.Flag) *pflag.Flag {
	return &pflag.Flag{
		Name:                f.Name,
		Shorthand:           f.Shorthand,
		Usage:               f.Usage,
		Value:               newBroadcastValue(f.Value),
		DefValue:            f.DefValue,
		Changed:             f.Changed,
		NoOptDefVal:         f.NoOptDefVal,
		Deprecated:          f.Deprecated,
		Hidden:              f.Hidden,
		ShorthandDeprecated: f.ShorthandDeprecated,
		Annotations:         f.Annotations,
	}
}

type broadcastValue []pflag.Value

func (bv broadcastValue) String() string {
	return bv[0].String()
}

func (bv broadcastValue) Set(s string) error {
	for _, v := range bv {
		if err := v.Set(s); err != nil {
			return err
		}
	}
	return nil
}

func (bv broadcastValue) Type() string {
	return bv[0].Type()
}

func newBroadcastValue(val pflag.Value) pflag.Value {
	return broadcastValue([]pflag.Value{val})
}
