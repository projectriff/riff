/*
 * Copyright 2018 The original author or authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package commands

import (
	"fmt"
	"strings"

	"unicode"

	"io"
	"text/template"

	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/util/validation"
)

// =============================================== Args related functions ==============================================

// ArgValidationConjunction returns a PositionalArgs validator that checks all provided validators in turn (all must pass).
func ArgValidationConjunction(validators ...cobra.PositionalArgs) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		for _, v := range validators {
			err := v(cmd, args)
			if err != nil {
				return err
			}
		}
		return nil
	}
}

// UpToDashDash returns a validator that will invoke the `delegate` validator, but only with args before the
// splitting `--`, if any
func UpToDashDash(delegate cobra.PositionalArgs) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		if cmd.ArgsLenAtDash() >= 0 {
			return delegate(cmd, args[0:cmd.ArgsLenAtDash()])
		} else {
			return delegate(cmd, args)
		}
	}
}

// PositionalArg is a function for validating a single argument
type PositionalArg func(cmd *cobra.Command, arg string) error

// AtPosition returns a PositionalArgs that applies the single valued validator to the i-th argument.
// The actual number of arguments is not checked by this function (use cobra's MinimumNArgs, ExactArgs, etc)
func AtPosition(i int, validator PositionalArg) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		return validator(cmd, args[i])
	}
}

func OptionalAtPosition(i int, validator PositionalArg) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		if len(args) > i {
			return validator(cmd, args[i])
		}
		return nil
	}
}

// KubernetesValidation turns a kubernetes-style validation function into a PositionalArg
func KubernetesValidation(k8s func(string) []string) PositionalArg {
	return func(cmd *cobra.Command, arg string) error {
		msgs := k8s(arg)
		if len(msgs) > 0 {
			return fmt.Errorf("%s", strings.Join(msgs, ", "))
		} else {
			return nil
		}
	}
}

func ValidName() PositionalArg {
	return KubernetesValidation(validation.IsDNS1123Subdomain)
}

func LabelArgs(cmd *cobra.Command, labels ...string) {
	if cmd.Annotations == nil {
		cmd.Annotations = make(map[string]string)
	}
	for i, label := range labels {
		cmd.Annotations[fmt.Sprintf("arg%d", i)] = label
	}
}

// =============================================== Flags related functions =============================================

type FlagsValidator func(cmd *cobra.Command) error

// CobraEFunction is the type of functions cobra expects for Run, PreRun, etc that can return an error.
type CobraEFunction func(cmd *cobra.Command, args []string) error

// FlagsValidatorAsCobraRunE allows a FlagsValidator to be used as a CobraEFunction (typically PreRunE())
func FlagsValidatorAsCobraRunE(validator FlagsValidator) CobraEFunction {
	return func(cmd *cobra.Command, args []string) error {
		return validator(cmd)
	}
}

// FlagsValidationConjunction returns a FlagsValidator validator that checks all provided validators in turn (all must pass).
func FlagsValidationConjunction(validators ...FlagsValidator) FlagsValidator {
	return func(cmd *cobra.Command) error {
		for _, v := range validators {
			err := v(cmd)
			if err != nil {
				return err
			}
		}
		return nil
	}
}

type FlagsMatcher interface {
	Evaluate(command *cobra.Command) bool
	Description() string
}

type flagsMatcher struct {
	eval func(command *cobra.Command) bool
	desc string
}

func (fm flagsMatcher) Evaluate(command *cobra.Command) bool {
	return fm.eval(command)
}

func (fm flagsMatcher) Description() string {
	return fm.desc
}

func Set(name string) FlagsMatcher {
	return flagsMatcher{
		eval: func(cmd *cobra.Command) bool {
			f := cmd.Flag(name)
			if f == nil {
				panic(fmt.Sprintf("Expected to find flag named %q in command %q", name, cmd.Use))
			}
			return f.Changed
		},
		desc: fmt.Sprintf("--%s is set", name),
	}
}

// FlagsDependency returns a validator that will evaluate the given delegate if the provided flag matcher returns true.
// Use to enforce scenarios such as "if --foo is set, then --bar must be set as well".
func FlagsDependency(matcher FlagsMatcher, delegate FlagsValidator) FlagsValidator {
	return func(cmd *cobra.Command) error {
		if matcher.Evaluate(cmd) {
			// Flag set. Delegate condition must HOLD
			err := delegate(cmd)
			if err != nil {
				return fmt.Errorf("when %v, %v", matcher.Description(), err)
			}
			return nil
		} else {
			// Flag not set. Don't check delegate.
			return nil
		}
	}
}

// AtLeastOneOf returns a FlagsValidator that asserts that at least one of the passed in flags is set.
func AtLeastOneOf(flagNames ...string) FlagsValidator {
	return func(cmd *cobra.Command) error {
		for _, f := range flagNames {
			flag := cmd.Flag(f)
			if flag == nil {
				panic(fmt.Sprintf("Expected to find flag named %q in command %q", f, cmd.Use))
			}
			if flag.Changed {
				return nil
			}
		}
		return fmt.Errorf("at least one of --%s must be set", strings.Join(flagNames, ", --"))
	}
}

// AtMostOneOf returns a FlagsValidator that asserts that at most one of the passed in flags is set.
func AtMostOneOf(flagNames ...string) FlagsValidator {
	return func(cmd *cobra.Command) error {
		set := 0
		for _, f := range flagNames {
			flag := cmd.Flag(f)
			if flag == nil {
				panic(fmt.Sprintf("Expected to find flag named %q in command %q", f, cmd.Use))
			}
			if flag.Changed {
				set++
			}
		}
		if set > 1 {
			return fmt.Errorf("at most one of --%s must be set", strings.Join(flagNames, ", --"))
		} else {
			return nil
		}
	}
}

// NotBlank returns a FlagsValidator that asserts that the given flag is not made entirely of space characters.
func NotBlank(flagName string) FlagsValidator {
	return func(cmd *cobra.Command) error {
		if strings.TrimSpace(cmd.Flag(flagName).Value.String()) == "" {
			return fmt.Errorf("flag --%s cannot be empty", flagName)
		}
		return nil
	}
}

// ExactlyOneOf returns a FlagsValidator that asserts that one and only one of the passed in flags is set.
func ExactlyOneOf(flagNames ...string) FlagsValidator {
	return FlagsValidationConjunction(
		AtLeastOneOf(flagNames...),
		AtMostOneOf(flagNames...),
	)
}

// =========================================== Usage related functions =================================================

func installAdvancedUsage(rootCmd *cobra.Command) {
	rootCmd.SetUsageFunc(func(c *cobra.Command) error {
		c.InitDefaultHelpFlag()
		err := tmpl(c.OutOrStderr(), c.UsageTemplate(), c)
		if err != nil {
			c.Println(err)
		}
		return err
	})
	rootCmd.SetUsageTemplate(`Usage:{{if .Runnable}}
  {{useline .}}{{end}}{{if .HasAvailableSubCommands}}
  {{.CommandPath}} [command]{{end}}{{if gt (len .Aliases) 0}}

Aliases:
  {{.NameAndAliases}}{{end}}{{if .HasExample}}

Examples:
{{.Example}}{{end}}{{if .HasAvailableSubCommands}}

Available Commands:{{range .Commands}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

Flags:
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableInheritedFlags}}

Global Flags:
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasHelpSubCommands}}

Additional help topics:{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
  {{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}

Use "{{.CommandPath}} [command] --help" for more information about a command.{{end}}
`)
}

func tmpl(w io.Writer, text string, data interface{}) error {
	t := template.New("top")
	templateFuncs := template.FuncMap{
		"trim":                    strings.TrimSpace,
		"trimRightSpace":          trimRightSpace,
		"trimTrailingWhitespaces": trimRightSpace,
		"rpad":                    rpad,
		"gt":                      cobra.Gt,
		"eq":                      cobra.Eq,
		"useline":                 useline,
	}
	t.Funcs(templateFuncs)
	template.Must(t.Parse(text))
	return t.Execute(w, data)
}

// rpad adds padding to the right of a string.
func rpad(s string, padding int) string {
	format := fmt.Sprintf("%%-%ds", padding)
	return fmt.Sprintf(format, s)
}

func trimRightSpace(s string) string {
	return strings.TrimRightFunc(s, unicode.IsSpace)
}

// useline returns the default cobra Useline() of a command, enhanced with markers for named arguments
func useline(c *cobra.Command) string {
	result := c.UseLine()
	flags := ""
	if strings.HasSuffix(result, " [flags]") {
		flags = " [flags]"
		result = result[0 : len(result)-len(flags)]
	}

	ok := true
	for i := 0; ok; i++ {
		info, found := c.Annotations[fmt.Sprintf("arg%d", i)]
		ok = found
		result += " " + info
	}

	return result + flags
}

// =========================================== General Cobra functions =================================================

// Visit applies the provided function f to the given command and its children, depth first.
// Exits as soon as an error occurs.
func Visit(cmd *cobra.Command, f func(c *cobra.Command) error) error {
	err := f(cmd)
	if err != nil {
		return err
	}
	for _, c := range cmd.Commands() {
		err := Visit(c, f)
		if err != nil {
			return err
		}
	}
	return nil
}

func PrintSuccessfulCompletion(cmd *cobra.Command) {
	fmt.Fprintf(cmd.OutOrStdout(), "\n%s completed successfully\n", cmd.CommandPath())
}

func PrintInterruptedCompletion(cmd *cobra.Command) {
	fmt.Fprintf(cmd.OutOrStdout(), "\n%s was interrupted\n", cmd.CommandPath())
}


func FindSubcommand(command *cobra.Command, names ...string) *cobra.Command {
	cmd, unmatchedArgs, err := command.Find(names)
	if err != nil || len(unmatchedArgs) > 0 {
		return nil
	}
	return cmd
}