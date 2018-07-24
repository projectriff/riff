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
	"fmt"

	projectriff_v1 "github.com/projectriff/riff/kubernetes-crds/pkg/apis/projectriff.io/v1alpha1"
	"github.com/projectriff/riff/riff-cli/pkg/templateutils"
	"github.com/spf13/cobra"
)

type Defaults struct {
	UserAccount string
	Force       bool
	DryRun      bool
	Push        bool
	Version     string
}

var DefaultValues = Defaults{
	UserAccount: "current OS user",
	Force:       false,
	DryRun:      false,
	Push:        false,
	Version:     "0.0.1",
}

const (
	initResult       = `generate the resource definitions using sensible defaults`
	initDefinition   = `Generate`
	createResult     = `create the resource definitions, and apply the resources, using sensible defaults`
	createDefinition = `Create`
)

const baseDescription = `{{.Process}} the function based on the function source code specified as the filename, using the name
and version specified for the function image repository and tag. 

For example, from a directory named 'square' containing a function 'square.js', you can simply type :

    riff {{.Command}}

to {{.Result}}.`

// Copied from cobra's default usage template to change the ambiguous "[command]" word to "[invoker]"
const CustomInvokerUsageTemplate = `Usage:{{if .Runnable}}
  {{.UseLine}}{{end}}{{if .HasAvailableSubCommands}}
  {{.CommandPath}} [invoker] [flags]{{end}}{{if gt (len .Aliases) 0}}

Aliases:
  {{.NameAndAliases}}{{end}}{{if .HasExample}}

Examples:
{{.Example}}{{end}}{{if .HasAvailableSubCommands}}

Available Invokers:{{range .Commands}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

Flags:
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableInheritedFlags}}

Global Flags:
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasHelpSubCommands}}

Additional help topics:{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
  {{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}

Use "{{.CommandPath}} [invoker] --help" for more information about an installed invoker.{{end}}
`


type LongVals struct {
	Process string
	Command string
	Result  string
}

func InitCmdLong() string {
	return templateCmdLong(baseDescription, LongVals{Process: initDefinition, Command: "init", Result: initResult})
}

func InitInvokerCmdLong(invoker projectriff_v1.Invoker) string {
	command := fmt.Sprintf("%s %s", "init", invoker.ObjectMeta.Name)
	return templateCmdLong(invoker.Spec.Doc, LongVals{Process: initDefinition, Command: command, Result: initResult})
}

func CreateCmdLong() string {
	return templateCmdLong(baseDescription, LongVals{Process: createDefinition, Command: "create", Result: createResult})
}

func CreateInvokerCmdLong(invoker projectriff_v1.Invoker) string {
	command := fmt.Sprintf("%s %s", "create", invoker.ObjectMeta.Name)
	return templateCmdLong(invoker.Spec.Doc, LongVals{Process: createDefinition, Command: command, Result: createResult})
}

func templateCmdLong(longDescrTmpl string, vals LongVals) string {
	longDescr, err := templateutils.Apply(longDescrTmpl, "longDescr", vals)
	if err != nil {
		panic(err)
	}
	return longDescr
}

// AliasFlagToSoleArg returns a cobra.PositionalArgs args validator that populates the given flag if it hasn't been yet,
// from an arg that must be set and be the only one. No args must be present if the flag has already been set.
func AliasFlagToSoleArg(flag string) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		f := cmd.Flag(flag)
		if len(args) > 0 {
			if len(args) == 1 {
				if !f.Changed {
					f.Value.Set(args[0])
				} else {
					return fmt.Errorf("value for %v has already been set via the --%v flag to '%v'. "+
						"Can't set it via an argument (to '%v') as well", flag, flag, f.Value.String(), args[0])
				}
			} else {
				return fmt.Errorf("command %v expects exactly one argument", cmd.Name())
			}
		}
		return nil
	}
}

func And(functions ...cobra.PositionalArgs) cobra.PositionalArgs {
	if len(functions) == 0 {
		return nil
	}
	return func(cmd *cobra.Command, args []string) error {
		for _, f := range functions {
			if err := f(cmd, args); err != nil {
				return err
			}
		}
		return nil
	}
}
