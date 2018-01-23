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

import "github.com/spf13/cobra"

/*
 * Runs a chain of commands
 */
func CommandChain(commands ... *cobra.Command) *cobra.Command {

	run := func(cmd *cobra.Command, args []string) {
		for _, command := range commands {
			if command.Run != nil {
				command.Run(cmd, args)
			}
		}
	}

	/*
	 * Composite command using RunE to fail fast. SilenceUsage enabled to supress usage message following
	 * run time errors. If it gets this far, the usage was likely correct.
	 */
	runE := func(cmd *cobra.Command, args []string) error {
		for _, command := range commands {
			if command.RunE != nil {
				err := command.RunE(command, args)
				if err != nil {
					cmd.SilenceUsage = true
					return err
				}
			}
		}
		return nil
	}

	preRun := func(cmd *cobra.Command, args []string) {
		for _, command := range commands {
			if command.PreRun != nil {
				command.PreRun(command, args)
			}
		}
	}

	preRunE := func(cmd *cobra.Command, args []string) error {
		for _, command := range commands {
			if command.PreRunE != nil {
				err := command.PreRunE(command, args)
				if err != nil {
					return err
				}
			}
		}
		return nil
	}

	postRun := func(cmd *cobra.Command, args []string) {
		for _, command := range commands {
			if command.PostRun != nil {
				command.PostRun(command, args)
			}
		}
	}

	postRunE := func(cmd *cobra.Command, args []string) error {
		for _, command := range commands {
			if command.PostRunE != nil {
				err := command.PostRunE(command, args)
				if err != nil {
					return err
				}
			}
		}
		return nil
	}

	persistentPreRun := func(cmd *cobra.Command, args []string) {

		for _, command := range commands {
			for ; command.Root() != nil && command.Root().PersistentPreRun != nil; {
				command = command.Root()
			}
			if command.PersistentPreRun != nil {
				command.PersistentPreRun(command, args)
			}
		}
	}

	persistentPreRunE := func(cmd *cobra.Command, args []string) error {
		for _, command := range commands {
			for ; command.Root() != nil && command.Root().PersistentPreRun != nil; {
				command = command.Root()
			}
			if command.PersistentPreRunE != nil {
				err := command.PersistentPreRunE(command, args)
				if err != nil {
					return err
				}
			}
		}
		return nil
	}

	persistentPostRun := func(cmd *cobra.Command, args []string) {

		for _, command := range commands {
			for ; command.Root() != nil && command.Root().PersistentPreRun != nil; {
				command = command.Root()
			}
			if command.PersistentPostRun != nil {
				command.PersistentPostRun(command, args)
			}
		}
	}

	persistentPostRunE := func(cmd *cobra.Command, args []string) error {
		for _, command := range commands {
			for ; command.Root() != nil && command.Root().PersistentPreRun != nil; {
				command = command.Root()
			}
			if command.PersistentPostRunE != nil {
				err := command.PersistentPostRunE(command, args)
				if err != nil {
					return err
				}
			}
		}
		return nil
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
	}

	return chain
}

