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

package commands

import (
	"context"
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/projectriff/cli/pkg/cli"
	corecommands "github.com/projectriff/cli/pkg/core/commands"
	knativecommands "github.com/projectriff/cli/pkg/knative/commands"
	streamingcommands "github.com/projectriff/cli/pkg/streaming/commands"
	"github.com/spf13/cobra"
)

// NewRootCommand wraps the riff command with flags and commands that should only be defined on
// the CLI root.
func NewRootCommand(ctx context.Context, c *cli.Config) *cobra.Command {
	cmd := NewRiffCommand(ctx, c)

	cmd.Use = c.Name
	cmd.DisableAutoGenTag = true

	// set version
	if !c.GitDirty {
		cmd.Version = fmt.Sprintf("%s (%s)", c.Version, c.GitSha)
	} else {
		cmd.Version = fmt.Sprintf("%s (%s, with local modifications)", c.Version, c.GitSha)
	}
	cmd.Flags().Bool("version", false, "display CLI version")

	// add root persistent flags
	cmd.PersistentFlags().StringVar(&c.ViperConfigFile, cli.StripDash(cli.ConfigFlagName), "", fmt.Sprintf("config `file` (default is $HOME/.%s.yaml)", c.Name))
	_ = cmd.MarkFlagFilename(cli.StripDash(cli.ConfigFlagName), "yaml", "yml")
	cmd.PersistentFlags().StringVar(&c.KubeConfigFile, cli.StripDash(cli.KubeConfigFlagName), "", "kubectl config `file` (default is $HOME/.kube/config)")
	cmd.PersistentFlags().StringVar(&c.KubeConfigFile, cli.StripDash(cli.KubeConfigFlagNameDeprecated), "", "kubectl config `file` (default is $HOME/.kube/config)")
	cmd.PersistentFlags().MarkDeprecated(cli.StripDash(cli.KubeConfigFlagNameDeprecated), fmt.Sprintf("renamed to %s", cli.KubeConfigFlagName))
	cmd.PersistentFlags().BoolVar(&color.NoColor, cli.StripDash(cli.NoColorFlagName), color.NoColor, "disable color output in terminals")

	// add runtimes
	runtimes := []struct {
		name    string
		command *cobra.Command
		doc     string
	}{{
		name:    cli.CoreRuntime,
		command: corecommands.NewCoreCommand(ctx, c),
		doc: strings.TrimSpace(`
The core runtime uses core Kubernetes resources like Deployment and Service to
expose the workload over HTTP.
`),
	}, {
		name:    cli.StreamingRuntime,
		command: streamingcommands.NewStreamingCommand(ctx, c),
		doc: strings.TrimSpace(`
The streaming runtime maps one or more input and output streams to a function.
`),
	}, {
		name:    cli.KnativeRuntime,
		command: knativecommands.NewKnativeCommand(ctx, c),
		doc: strings.TrimSpace(`
The Knative runtime uses Knative Serving to expose the workload over HTTP with
zero-to-n autoscaling and managed ingress.
`),
	}}
	for _, runtime := range runtimes {
		if c.Runtimes[runtime.name] {
			cmd.Long = cmd.Long + "\n\n" + runtime.doc
		} else {
			runtime.command.Hidden = true
		}
		cmd.AddCommand(runtime.command)
	}

	// add root-only commands
	cmd.AddCommand(NewCompletionCommand(ctx, c))
	cmd.AddCommand(NewDocsCommand(ctx, c))
	cmd.AddCommand(NewDoctorCommand(ctx, c))

	// override usage template to add arguments
	cmd.SetUsageTemplate(strings.ReplaceAll(cmd.UsageTemplate(), "{{.UseLine}}", "{{useLine .}}"))
	cobra.AddTemplateFunc("useLine", func(cmd *cobra.Command) string {
		result := cmd.UseLine()
		flags := ""
		if strings.HasSuffix(result, " [flags]") {
			flags = " [flags]"
			result = result[0 : len(result)-len(flags)]
		}
		return result + cli.FormatArgs(cmd) + flags
	})

	return cmd
}
