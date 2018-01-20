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

package cmd

import (
	"github.com/spf13/cobra"
	"github.com/projectriff/riff-cli/pkg/osutils"
	"github.com/spf13/pflag"
)

func createInitOptionFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().StringP("name", "n", "", "the functionName of the function (defaults to the functionName of the current directory)")
	cmd.PersistentFlags().StringP("version", "v", "0.0.1", "the version of the function (defaults to 0.0.1)")
	cmd.PersistentFlags().StringP("filepath", "f", "", "path or directory to be used for the function resources, if a file is specified then the file's directory will be used (defaults to the current directory)")
	cmd.PersistentFlags().StringP("protocol", "p", "", "the protocol to use for function invocations (defaults to 'stdio' for shell and python, to 'http' for java and node)")
	cmd.PersistentFlags().StringP("input", "i", "", "the functionName of the input topic (defaults to function functionName)")
	cmd.PersistentFlags().StringP("output", "o", "", "the functionName of the output topic (optional)")
	cmd.PersistentFlags().StringP("artifact", "a", "", "path to the function artifact, source code or jar file")
	cmd.PersistentFlags().StringP("riff-version", "", RIFF_VERSION, "the version of riff to use when building containers")
	cmd.PersistentFlags().StringP("useraccount", "u", osutils.GetCurrentUsername(), "the Docker user account to be used for the image repository (defaults to current OS username")
}

func loadInitOptions(flagset pflag.FlagSet) InitOptions {
	opts := InitOptions{}
	opts.functionName, _ 	= flagset.GetString("name")
	opts.version, _ 		= flagset.GetString("version")
	opts.functionPath, _ 	= flagset.GetString("filepath")
	opts.protocol, _ 		= flagset.GetString("protocol")
	opts.input, _ 			= flagset.GetString("input")
	opts.output, _ 			= flagset.GetString("output")
	opts.artifact, _ 		= flagset.GetString("artifact")
	opts.riffVersion, _ 	= flagset.GetString("riff-version")
	opts.userAccount, _ 	= flagset.GetString("useraccount")
	return opts
}

/*
 * Runs a chain of commands
 */
func commandChain(commands ... *cobra.Command) *cobra.Command {

	run := func(cmd *cobra.Command, args []string) {
		for _, command := range commands {
			if command.Run != nil {
				command.Run(cmd, args)
			}
		}
	}

	runE := func(cmd *cobra.Command, args []string) error {
		for _, command := range commands {
			if command.RunE != nil {
				err := command.RunE(cmd, args)
				if err != nil {
					return err
				}
			}
		}
		return nil
	}

	preRun := func(cmd *cobra.Command, args []string) {
		for _, command := range commands {
			if command.PreRun != nil {
				command.PreRun(cmd, args)
			}
		}
	}

	preRunE := func(cmd *cobra.Command, args []string) error {
		for _, command := range commands {
			if command.PreRunE != nil {
				err := command.PreRunE(cmd, args)
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
				command.PostRun(cmd, args)
			}
		}
	}

	postRunE := func(cmd *cobra.Command, args []string) error {
		for _, command := range commands {
			if command.PostRunE != nil {
				err := command.PostRunE(cmd, args)
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
				command.PersistentPreRun(cmd, args)
			}
		}
	}

	persistentPreRunE := func(cmd *cobra.Command, args []string) error {
		for _, command := range commands {
			for ; command.Root() != nil && command.Root().PersistentPreRun != nil; {
				command = command.Root()
			}
			if command.PersistentPreRunE != nil {
				err := command.PersistentPreRunE(cmd, args)
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
				command.PersistentPostRun(cmd, args)
			}
		}
	}

	persistentPostRunE := func(cmd *cobra.Command, args []string) error {
		for _, command := range commands {
			for ; command.Root() != nil && command.Root().PersistentPreRun != nil; {
				command = command.Root()
			}
			if command.PersistentPostRunE != nil {
				err := command.PersistentPostRunE(cmd, args)
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
