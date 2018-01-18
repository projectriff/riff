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
)

func createInitOptionFlags(cmd *cobra.Command , options *InitOptions) {
	cmd.PersistentFlags().StringVarP(&options.functionName, "name", "n", "", "the functionName of the function (defaults to the functionName of the current directory)")
	cmd.PersistentFlags().StringVarP(&options.version, "version", "v", "0.0.1", "the version of the function (defaults to 0.0.1)")
	cmd.PersistentFlags().StringVarP(&options.functionPath, "filepath", "f", "", "path or directory to be used for the function resources, if a file is specified then the file's directory will be used (defaults to the current directory)")
	cmd.PersistentFlags().StringVarP(&options.protocol, "protocol", "p", "", "the protocol to use for function invocations (defaults to 'stdio' for shell and python, to 'http' for java and node)")
	cmd.PersistentFlags().StringVarP(&options.input, "input", "i", "", "the functionName of the input topic (defaults to function functionName)")
	cmd.PersistentFlags().StringVarP(&options.output, "output", "o", "", "the functionName of the output topic (optional)")
	cmd.PersistentFlags().StringVarP(&options.artifact, "artifact", "a", "", "path to the function artifact, source code or jar file")
}

func createBuildOptionFlags(cmd *cobra.Command , options *BuildOptions) {
	createCmd.PersistentFlags().StringVarP(&options.userAccount, "useraccount", "u", osutils.GetCurrentUsername(), "the Docker user account to be used for the image repository (defaults to current OS username")
	createCmd.PersistentFlags().StringVarP(&options.riffVersion, "riff-version", "", "0.0.1", "the version of riff to use when building containers")
	createCmd.PersistentFlags().BoolVarP(&options.push, "push", "", false, "push the image to Docker registry")
}

/*
 * Runs a chain of commands
 */
func CommandChain(commands... *cobra.Command)  *cobra.Command {

	run := func(cmd *cobra.Command, args []string) {
		for _,command := range commands {
			if command.Run != nil {
				command.Run(cmd, args)
			}
		}
	}

	runE := func(cmd *cobra.Command, args []string) error {
		for _,command := range commands {
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
		for _,command := range commands {
			if command.PreRun != nil {
				command.PreRun(cmd, args)
			}
		}
	}

	preRunE := func(cmd *cobra.Command, args []string) error {
		for _,command := range commands {
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
		for _,command := range commands {
			if command.PostRun != nil {
				command.PostRun(cmd, args)
			}
		}
	}

	postRunE := func(cmd *cobra.Command, args []string) error {
		for _,command := range commands {
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
		for _,command := range commands {
			if command.PersistentPreRun != nil {
				command.PersistentPreRun(cmd, args)
			}
		}
	}

	persistentPreRunE := func(cmd *cobra.Command, args []string) error {
		for _,command := range commands {
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
		for _,command := range commands {
			if command.PersistentPostRun != nil {
				command.PersistentPostRun(cmd, args)
			}
		}
	}

	persistentPostRunE := func(cmd *cobra.Command, args []string) error {
		for _,command := range commands {
			if command.PersistentPostRunE != nil {
				err := command.PersistentPostRunE(cmd, args)
				if err != nil {
					return err
				}
			}
		}
		return nil
	}



	var chain = &cobra.Command {
		Run: run,
		RunE: runE,
		PreRun:preRun,
		PreRunE:preRunE,
		PostRun:postRun,
		PostRunE:postRunE,
		PersistentPreRun:persistentPreRun,
		PersistentPreRunE:persistentPreRunE,
		PersistentPostRun:persistentPostRun,
		PersistentPostRunE:persistentPostRunE,
	}

	return chain
}

