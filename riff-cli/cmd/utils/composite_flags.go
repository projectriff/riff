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
	"github.com/projectriff/riff-cli/pkg/osutils"
	"github.com/spf13/pflag"
	"github.com/projectriff/riff-cli/pkg/options"
	"github.com/projectriff/riff-cli/global"
	"github.com/spf13/cobra"
	"github.com/projectriff/riff-cli/cmd/opts"
	"github.com/spf13/viper"
)

type Defaults struct {
	riffVersion string
	userAccount string
	namespace   string
	force       bool
	dryRun      bool
	push        bool
	version     string
}

var defaults = Defaults{
	riffVersion: global.RIFF_VERSION,
	userAccount: "current OS user",
	namespace:   "default",
	force:       false,
	dryRun:      false,
	push:        false,
	version:     "0.0.1",
}

func CreateInitFlags(flagset *pflag.FlagSet) {
	setVersionFlag(flagset)
	setNameFlag(flagset)
	setFilePathFlag(flagset)
	setProtocolFlag(flagset)
	setInputFlag(flagset)
	setOutputFlag(flagset)
	setArtifactFlag(flagset)
	setRiffVersionFlag(flagset)
	setUserAccountFlag(flagset)
	setForceFlag(flagset)
	setDryRunFlag(flagset)
}

func CreateBuildFlags(flagset *pflag.FlagSet) {
	setNameFlag(flagset)
	setFilePathFlag(flagset)
	setVersionFlag(flagset)
	setRiffVersionFlag(flagset)
	setDryRunFlag(flagset)
	setPushFlag(flagset)
	setUserAccountFlag(flagset)
}

func CreateApplyFlags(flagset *pflag.FlagSet) {
	setFilePathFlag(flagset)
	setNamespaceFlag(flagset)
	setDryRunFlag(flagset)
}

func CreateDeleteFlags(flagset *pflag.FlagSet) {
	setFilePathFlag(flagset)
	setNameFlag(flagset)
	setNamespaceFlag(flagset)
	setDryRunFlag(flagset)
	flagset.Bool("all", false, "delete all resources including topics, not just the function resource")
}

func MergeInitOptions(flagset pflag.FlagSet, opts *options.InitOptions) {
	if opts.FunctionName == "" {
		opts.FunctionName, _ = flagset.GetString("name")
	}
	if opts.Version == "" {
		opts.Version, _ = flagset.GetString("version")
	}
	if opts.FilePath == "" {
		opts.FilePath, _ = flagset.GetString("filepath")
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
		opts.RiffVersion = getRiffVersionWithGlobalOverride(flagset)
	}
	if opts.UserAccount == "" {
		opts.UserAccount = getUseraccountWithOverride("useraccount", flagset)
	}
	if opts.DryRun == false {
		opts.DryRun, _ = flagset.GetBool("dry-run")
	}
	if opts.Force == false {
		opts.Force, _ = flagset.GetBool("force")
	}
}

func MergeBuildOptions(flagset pflag.FlagSet, opts *options.CreateOptions) {
	if opts.FunctionName == "" {
		opts.FunctionName, _ = flagset.GetString("name")
	}
	if opts.Version == "" {
		opts.Version, _ = flagset.GetString("version")
	}
	if opts.FilePath == "" {
		opts.FilePath, _ = flagset.GetString("filepath")
	}
	if opts.RiffVersion == "" {
		opts.RiffVersion = getRiffVersionWithGlobalOverride(flagset)
	}
	if opts.UserAccount == "" {
		opts.UserAccount = getUseraccountWithOverride("useraccount", flagset)
	}
	if opts.DryRun == false {
		opts.DryRun, _ = flagset.GetBool("dry-run")
	}
	if opts.Push == false {
		opts.Push, _ = flagset.GetBool("push")
	}
}

func MergeApplyOptions(flagset pflag.FlagSet, opts *options.CreateOptions) {
	if opts.FilePath == "" {
		opts.FilePath, _ = flagset.GetString("filepath")
	}
	if opts.Namespace == "" {
		opts.Namespace = GetStringValueWithOverride("namespace", flagset)
	}
	if opts.DryRun == false {
		opts.DryRun, _ = flagset.GetBool("dry-run")
	}
}

func MergeDeleteOptions(flagset pflag.FlagSet, opts *options.DeleteAllOptions) {
	if opts.FunctionName == "" {
		opts.FunctionName, _ = flagset.GetString("name")
	}
	if opts.FilePath == "" {
		opts.FilePath, _ = flagset.GetString("filepath")
	}
	if opts.Namespace == "" {
		opts.Namespace = GetStringValueWithOverride("namespace", flagset)
	}
	if opts.DryRun == false {
		opts.DryRun, _ = flagset.GetBool("dry-run")
	}
	if opts.All == false {
		opts.All, _ = flagset.GetBool("all")
	}
}

func GetHandler(cmd *cobra.Command) string {
	if opts.Handler == "" {
		opts.Handler, _ = cmd.Flags().GetString("handler")
	}
	return opts.Handler
}


func setNameFlag(flagset *pflag.FlagSet) {
	if !flagDefined(flagset, "name") {
		flagset.StringP("name", "n", "", "the name of the function (defaults to the name of the current directory)")
	}
}

func setVersionFlag(flagset *pflag.FlagSet) {
	if !flagDefined(flagset, "version") {
		flagset.StringP("version", "v", defaults.version, "the version of the function image")
	}
}

func setFilePathFlag(flagset *pflag.FlagSet) {
	if !flagDefined(flagset, "filepath") {
		flagset.StringP("filepath", "f", "", "path or directory used for the function resources (defaults to the current directory)")
	}
}

func setNamespaceFlag(flagset *pflag.FlagSet) {
	if !flagDefined(flagset, "namespace") {
		flagset.String("namespace", defaults.namespace, "the namespace used for the deployed resources")
	}
}

func setDryRunFlag(flagset *pflag.FlagSet) {
	if !flagDefined(flagset, "dry-run") {
		flagset.Bool("dry-run", defaults.dryRun, "print generated function artifacts content to stdout only")
	}
}

func setRiffVersionFlag(flagset *pflag.FlagSet) {
	if !flagDefined(flagset, "riff-version") {
		flagset.StringP("riff-version", "", defaults.riffVersion, "the version of riff to use when building containers")
	}
}

func setUserAccountFlag(flagset *pflag.FlagSet) {
	if !flagDefined(flagset, "useraccount") {
		flagset.StringP("useraccount", "u", defaults.userAccount, "the Docker user account to be used for the image repository")
	}
}

func setProtocolFlag(flagset *pflag.FlagSet) {
	if !flagDefined(flagset, "protocol") {
		flagset.StringP("protocol", "p", "", "the protocol to use for function invocations")
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
		flagset.Bool("force", defaults.force, "overwrite existing functions artifacts")
	}
}

func setPushFlag(flagset *pflag.FlagSet) {
	if !flagDefined(flagset, "push") {
		flagset.BoolP("push", "", defaults.push, "push the image to Docker registry")
	}
}

func flagDefined(flagset *pflag.FlagSet, name string) bool {
	return flagset.Lookup(name) != nil
}

func getRiffVersionWithGlobalOverride(flagset pflag.FlagSet) string {
	val, _ := flagset.GetString("riff-version")
	if flagset.Changed("riff-version") {
		return val
	}
	return global.RIFF_VERSION
}

func getUseraccountWithOverride(name string, flagset pflag.FlagSet) string {
	userAcct := GetStringValueWithOverride("useraccount", flagset)
	if userAcct == defaults.userAccount {
		userAcct = osutils.GetCurrentUsername()
	}
	return userAcct
}

func GetStringValueWithOverride(name string, flagset pflag.FlagSet) string {
	viperVal := viper.GetString(name)
	val, _ := flagset.GetString(name)
	if flagset.Changed(name) || viperVal == "" {
		return val
	}
	return viperVal
}
