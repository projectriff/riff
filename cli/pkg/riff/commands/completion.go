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

	"github.com/projectriff/cli/pkg/cli"
	"github.com/spf13/cobra"
)

type CompletionOptions struct {
	Shell string
}

var (
	_ cli.Validatable = (*CompletionOptions)(nil)
	_ cli.Executable  = (*CompletionOptions)(nil)
)

func (opts *CompletionOptions) Validate(ctx context.Context) cli.FieldErrors {
	errs := cli.FieldErrors{}

	if opts.Shell == "" {
		errs = errs.Also(cli.ErrMissingField(cli.ShellFlagName))
	} else if opts.Shell != "bash" && opts.Shell != "zsh" {
		errs = errs.Also(cli.ErrInvalidValue(opts.Shell, cli.ShellFlagName))
	}

	return errs
}

func (opts *CompletionOptions) Exec(ctx context.Context, c *cli.Config) error {
	cmd := cli.CommandFromContext(ctx)
	switch opts.Shell {
	case "bash":
		return cmd.Root().GenBashCompletion(c.Stdout)
	case "zsh":
		return cmd.Root().GenZshCompletion(c.Stdout)
	}
	// protected by opts.Validate()
	panic("invalid shell: " + opts.Shell)
}

func (opts *CompletionOptions) MakeBashCompletion(c *cli.Config) string {
	return `
__` + c.Name + `_override_flag_list=(--kubeconfig --namespace -n)
__` + c.Name + `_override_flags()
{
	local ${__` + c.Name + `_override_flag_list[*]##*-} two_word_of of var
	for w in "${words[@]}"; do
		if [ -n "${two_word_of}" ]; then
			eval "${two_word_of##*-}=\"${two_word_of}=\${w}\""
			two_word_of=
			continue
		fi
		for of in "${__` + c.Name + `_override_flag_list[@]}"; do
			case "${w}" in
				${of}=*)
					eval "${of##*-}=\"${w}\""
					;;
				${of})
					two_word_of="${of}"
					;;
			esac
		done
	done
	for var in "${__` + c.Name + `_override_flag_list[@]##*-}"; do
		if eval "test -n \"\$${var}\""; then
			eval "echo -n \${${var}}' '"
		fi
	done
}

__` + c.Name + `_list_namespaces()
{
	local template
	template="{{ range .items }}{{ .metadata.name }} {{ end }}"
	local kubectl_out
	# TODO decouple from kubectl
	if kubectl_out=$(kubectl get $(__` + c.Name + `_override_flags) -o template --template="${template}" namespace 2>/dev/null); then
		COMPREPLY=( $( compgen -W "${kubectl_out}" -- "$cur" ) )
	fi
}

__` + c.Name + `_list_knative_configurations()
{
	local template
	template="{{ range .items }}{{ .metadata.name }} {{ end }}"
	local kubectl_out
	# TODO decouple from kubectl
	if kubectl_out=$(kubectl get $(__` + c.Name + `_override_flags) -o template --template="${template}" configurations.serving.knative.dev 2>/dev/null); then
		COMPREPLY=( $( compgen -W "${kubectl_out}" -- "$cur" ) )
	fi
}

__` + c.Name + `_list_knative_services()
{
	local template
	template="{{ range .items }}{{ .metadata.name }} {{ end }}"
	local kubectl_out
	# TODO decouple from kubectl
	if kubectl_out=$(kubectl get $(__` + c.Name + `_override_flags) -o template --template="${template}" services.serving.knative.dev 2>/dev/null); then
		COMPREPLY=( $( compgen -W "${kubectl_out}" -- "$cur" ) )
	fi
}

__` + c.Name + `_list_streaming_gateways()
{
	local template
	template="{{ range .items }}{{ .metadata.name }} {{ end }}"
	local kubectl_out
	# TODO decouple from kubectl
	if kubectl_out=$(kubectl get $(__` + c.Name + `_override_flags) -o template --template="${template}" gateways.streaming.projectriff.io 2>/dev/null); then
		COMPREPLY=( $( compgen -W "${kubectl_out}" -- "$cur" ) )
	fi
}

__` + c.Name + `_list_functions()
{
	__` + c.Name + `_list_resource 'function list'
}

__` + c.Name + `_list_containers()
{
	__` + c.Name + `_list_resource 'container list'
}

__` + c.Name + `_list_applications()
{
	__` + c.Name + `_list_resource 'application list'
}

__` + c.Name + `_list_resource()
{
	__` + c.Name + `_debug "listing $1"
	local ` + c.Name + `_output out
	if ` + c.Name + `_output=$(riff $1 $(__` + c.Name + `_override_flags) 2>/dev/null); then
		out=($(echo "${` + c.Name + `_output}" | awk 'NR>1 {print $1}'))
		COMPREPLY=( $( compgen -W "${out[*]}" -- "$cur" ) )
	fi
}

__` + c.Name + `_ingress_policy()
{
	COMPREPLY=( $( compgen -W "ClusterLocal External" -- "$cur" ) )
}

__` + c.Name + `_custom_func() {
	case ${last_command} in
		` + c.Name + `_application_delete | ` + c.Name + `_application_status | ` + c.Name + `_application_tail)
			__` + c.Name + `_list_resource 'application list'
			return
			;;
		` + c.Name + `_container_delete | ` + c.Name + `_container_status)
			__` + c.Name + `_list_resource 'container list'
			return
			;;
		` + c.Name + `_core_deployer_delete | ` + c.Name + `_core_deployer_status | ` + c.Name + `_core_deployer_tail)
			__` + c.Name + `_list_resource 'core deployer list'
			return
			;;
		` + c.Name + `_credential_delete)
			__` + c.Name + `_list_resource 'credential list'
			return
			;;
		` + c.Name + `_function_delete | ` + c.Name + `_function_status | ` + c.Name + `_function_tail)
			__` + c.Name + `_list_resource 'function list'
			return
			;;
		` + c.Name + `_knative_deployer_delete | ` + c.Name + `_knative_deployer_status | ` + c.Name + `_knative_deployer_tail)
			__` + c.Name + `_list_resource 'knative deployer list'
			return
			;;
		` + c.Name + `_knative_adapter_delete | ` + c.Name + `_knative_adapter_status)
			__` + c.Name + `_list_resource 'knative adapter list'
			return
			;;
		` + c.Name + `_streaming_inmemory-gateway_delete | ` + c.Name + `_streaming_inmemory-gateway_status)
			__` + c.Name + `_list_resource 'streaming inmemory-gateway list'
			return
			;;
		` + c.Name + `_streaming_kafka-gateway_delete | ` + c.Name + `_streaming_kafka-gateway_status)
			__` + c.Name + `_list_resource 'streaming kafka-gateway list'
			return
			;;
		` + c.Name + `_streaming_pulsar-gateway_delete | ` + c.Name + `_streaming_pulsar-gateway_status)
			__` + c.Name + `_list_resource 'streaming pulsar-gateway list'
			return
			;;
		` + c.Name + `_streaming_processor_delete | ` + c.Name + `_streaming_processor_status | ` + c.Name + `_streaming_processor_tail)
			__` + c.Name + `_list_resource 'streaming processor list'
			return
			;;
		` + c.Name + `_streaming_stream_delete | ` + c.Name + `_streaming_stream_status)
			__` + c.Name + `_list_resource 'streaming stream list'
			return
			;;
		*)
			;;
	esac
}
`
}

func NewCompletionCommand(ctx context.Context, c *cli.Config) *cobra.Command {
	opts := &CompletionOptions{}

	cmd := &cobra.Command{
		Use:   "completion",
		Short: "generate shell completion script",
		Long: strings.TrimSpace(`
Generate the completion script for your shell. The script is printed to stdout
and needs to be placed in the appropriate directory on your system.
`),
		Example: strings.Join([]string{
			fmt.Sprintf("%s completion", c.Name),
			fmt.Sprintf("%s completion %s zsh", c.Name, cli.ShellFlagName),
		}, "\n"),
		PreRunE: cli.Sequence(
			func(cmd *cobra.Command, args []string) error {
				cmd.Root().BashCompletionFunction = opts.MakeBashCompletion(c)
				return nil
			},
			cli.ValidateOptions(ctx, opts),
		),
		RunE: cli.ExecOptions(ctx, c, opts),
	}

	cmd.Flags().StringVar(&opts.Shell, cli.StripDash(cli.ShellFlagName), "bash", "`shell` to generate completion for: bash or zsh")

	return cmd
}
