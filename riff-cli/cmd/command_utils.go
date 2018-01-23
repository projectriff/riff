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
	"bytes"
	"text/template"
	"github.com/projectriff/riff-cli/pkg/options"
)

type Defaults struct {
	riffVersion string
	userAccount string
	force       bool
	dryRun      bool
	push        bool
	version     string
}

var DEFAULTS = Defaults{
	riffVersion: RIFF_VERSION,
	userAccount: osutils.GetCurrentUsername(),
	force:       false,
	dryRun:      false,
	push:        false,
	version:     "0.0.1",
}

func createInitFlags(flagset *pflag.FlagSet) {
	setVersionFlag(flagset)
	setNameFlag(flagset)
	setFilePathFlag(flagset)
	setRiffVersionFlag(flagset)
	setProtocolFlag(flagset)
	setInputFlag(flagset)
	setOutputFlag(flagset)
	setArtifactFlag(flagset)
	setRiffVersionFlag(flagset)
	setUserAccountFlag(flagset)
	setForceFlag(flagset)
	setDryRunFlag(flagset)
}

func createBuildFlags(flagset *pflag.FlagSet) {
	setNameFlag(flagset)
	setFilePathFlag(flagset)
	setVersionFlag(flagset)
	setRiffVersionFlag(flagset)
	setDryRunFlag(flagset)
	setPushFlag(flagset)
	setUserAccountFlag(flagset)
}

func createApplyFlags(flagset *pflag.FlagSet) {
	setFilePathFlag(flagset)
	setDryRunFlag(flagset)
}

func mergeInitOptions(flagset pflag.FlagSet, opts *options.InitOptions) {
	if opts.FunctionName == "" {
		opts.FunctionName, _ = flagset.GetString("name")
	}
	if opts.Version == "" {
		opts.Version, _ = flagset.GetString("version")
	}
	if opts.FunctionPath == "" {
		opts.FunctionPath, _ = flagset.GetString("filepath")
	}
	if opts.Protocol == "" {
		opts.Protocol, _ = flagset.GetString("protocol")
	}
	if opts.Input == "" {
		opts.Input, _ = flagset.GetString("input")
	}
	if opts.Output == "" {
		opts.Output, _ = flagset.GetString("output")
	}
	if opts.Artifact == "" {
		opts.Artifact, _ = flagset.GetString("artifact")
	}
	if opts.RiffVersion == "" {
		opts.RiffVersion, _ = flagset.GetString("riff-version")
	}
	if opts.UserAccount == "" {
		opts.UserAccount, _ = flagset.GetString("useraccount")
	}
	if opts.DryRun == false {
		opts.DryRun, _ = flagset.GetBool("dry-run")
	}
	if opts.Force == false {
		opts.Force, _ = flagset.GetBool("force")
	}
}

func mergeBuildOptions(flagset pflag.FlagSet, opts *options.CreateOptions) {
	if opts.FunctionName == "" {
		opts.FunctionName, _ = flagset.GetString("name")
	}
	if opts.Version == "" {
		opts.Version, _ = flagset.GetString("version")
	}
	if opts.FunctionPath == "" {
		opts.FunctionPath, _ = flagset.GetString("filepath")
	}
	if opts.RiffVersion == "" {
		opts.RiffVersion, _ = flagset.GetString("riff-version")
	}
	if opts.UserAccount == "" {
		opts.UserAccount, _ = flagset.GetString("useraccount")
	}
	if opts.DryRun == false {
		opts.DryRun, _ = flagset.GetBool("dry-run")
	}
	if opts.Push == false {
		opts.Push, _ = flagset.GetBool("push")
	}
}

func mergeApplyOptions(flagset pflag.FlagSet, opts *options.InitOptions) {
	if opts.FunctionPath == "" {
		opts.FunctionPath, _ = flagset.GetString("filepath")
	}
	if opts.DryRun == false {
		opts.DryRun, _ = flagset.GetBool("dry-run")
	}
}

func setNameFlag(flagset *pflag.FlagSet) {
	if !flagDefined(flagset, "name") {
		flagset.StringP("name", "n", "", "the name of the function (defaults to the functionName of the current directory)")
	}
}

func setVersionFlag(flagset *pflag.FlagSet) {
	if !flagDefined(flagset, "version") {
		flagset.StringP("version", "v", DEFAULTS.version, "the version of the function image")
	}
}

func setFilePathFlag(flagset *pflag.FlagSet) {
	if !flagDefined(flagset, "filepath") {
		flagset.StringP("filepath", "f", "", "path or directory to be used for the function resources, if a file is specified then the file's directory will be used (defaults to the current directory)")
	}
}

func setDryRunFlag(flagset *pflag.FlagSet) {
	if !flagDefined(flagset, "dry-run") {
		flagset.Bool("dry-run", DEFAULTS.dryRun, "print generated function artifacts content to stdout only")
	}
}

func setRiffVersionFlag(flagset *pflag.FlagSet) {
	if !flagDefined(flagset, "riff-version") {
		flagset.StringP("riff-version", "", DEFAULTS.riffVersion, "the version of riff to use when building containers")
	}
}

func setUserAccountFlag(flagset *pflag.FlagSet) {
	if !flagDefined(flagset, "useraccount") {
		flagset.StringP("useraccount", "u", DEFAULTS.userAccount, "the Docker user account to be used for the image repository (defaults to current OS username)")
	}
}

func setProtocolFlag(flagset *pflag.FlagSet) {
	if !flagDefined(flagset, "protocol") {
		flagset.StringP("protocol", "p", "", "the protocol to use for function invocations (defaults to 'stdio' for shell and python, to 'http' for java and node)")
	}
}

func setInputFlag(flagset *pflag.FlagSet) {
	if !flagDefined(flagset, "input") {
		flagset.StringP("input", "i", "", "the name of the input topic (defaults to function name)")
	}
}
func setOutputFlag(flagset *pflag.FlagSet) {
	if !flagDefined(flagset, "output") {
		flagset.StringP("output", "o", "", "the name of the output topic (optional)")
	}
}

func setArtifactFlag(flagset *pflag.FlagSet) {
	if !flagDefined(flagset, "artifact") {
		flagset.StringP("artifact", "a", "", "path to the function artifact, source code or jar file")
	}
}

func setForceFlag(flagset *pflag.FlagSet) {
	if !flagDefined(flagset, "force") {
		flagset.Bool("force", DEFAULTS.force, "overwrite existing functions artifacts")
	}
}

func setPushFlag(flagset *pflag.FlagSet) {
	if !flagDefined(flagset, "push") {
		flagset.BoolP("push", "", DEFAULTS.push, "push the image to Docker registry")
	}
}

func flagDefined(flagset *pflag.FlagSet, name string) bool {
	return flagset.Lookup(name) != nil
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

type LongVals struct {
	Process string
	Command string
	Result  string
}

func createCmdLong(longDescr string, vals LongVals) string {
	tmpl, err := template.New("longDescr").Parse(longDescr)
	if err != nil {
		panic(err)
	}

	var tpl bytes.Buffer
	err = tmpl.Execute(&tpl, vals)
	if err != nil {
		panic(err)
	}

	return tpl.String()
}
