/*
 * Copyright 2018-2019 The original author or authors
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
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/frioux/shellquote"
	"github.com/projectriff/riff/pkg/core"
	"github.com/projectriff/riff/pkg/core/tasks"
	"github.com/projectriff/riff/pkg/env"
	projectriffv1alpha1 "github.com/projectriff/system/pkg/apis/projectriff/v1alpha1"
	"github.com/spf13/cobra"
	v1 "k8s.io/api/core/v1"
)

const (
	applicationCreateApplicationNameIndex = iota
	applicationCreateNumberOfArgs
)

const (
	applicationUpdateApplicationNameIndex = iota
	applicationUpdateNumberOfArgs
)

const (
	applicationStatusApplicationNameIndex = iota
	applicationStatusNumberOfArgs
)

const (
	applicationBuildNumberOfArgs = iota
)

const (
	applicationListNumberOfArgs = iota
)

const (
	applicationInvokeServiceNameIndex = iota
	applicationInvokeServicePathIndex
	applicationInvokeMaxNumberOfArgs
)

const (
	applicationDeleteNameStartIndex = iota
	applicationDeleteMinNumberOfArgs
)

func Application() *cobra.Command {
	return &cobra.Command{
		Use:   "application",
		Short: "Interact with application related resources",
	}
}

func ApplicationCreate(buildpackBuilder core.Builder, fcTool *core.Client) *cobra.Command {
	createApplicationOptions := core.CreateApplicationOptions{}

	command := &cobra.Command{
		Use:   "create",
		Short: "Create a new application resource",
		Long: "Create a new application resource from the content of the provided Git repo/revision or local source.\n\n" +
			"Images will be pushed to the registry specified in the image name. If a default image prefix was specified during namespace init, the image flag is optional. The application name is combined with the default prefix to define the image. Instead of using the application name, a custom repository can be specified with the image set like `--image _/custom-name` which would resolve to `docker.io/example/custom-name`.\n\n" +
			"From then on you can use the sub-commands for the `service` command to interact with the service created for the application.\n\n" +
			envFromLongDesc + "\n",
		Example: `  ` + env.Cli.Name + ` application create square --git-repo https://github.com/acme/square --artifact square.js --image acme/square --namespace joseph-ns
  ` + env.Cli.Name + ` application create tweets-logger --git-repo https://github.com/acme/tweets --image acme/tweets-logger:1.0.0`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			validator := FlagsValidatorAsCobraRunE(AtLeastOneOf("git-repo", "local-path"))
			err := validator(cmd, args)
			if err != nil {
				return err
			}

			if createApplicationOptions.Image == "" || strings.HasPrefix(createApplicationOptions.Image, "_") {
				prefix, err := (*fcTool).DefaultBuildImagePrefix(createApplicationOptions.Namespace)
				if err != nil {
					return fmt.Errorf("unable to default image: %s", err)
				}
				if prefix == "" {
					if createApplicationOptions.Image == "" {
						return fmt.Errorf("required flag(s) \"image\" not set, this flag is optional if --image-prefix is specified during namespace init")
					}
					return fmt.Errorf("--image flag must include a repository, the image prefix was not set during namespace init")
				}

				fnName := args[applicationCreateApplicationNameIndex]
				if createApplicationOptions.Image == "" {
					// combine prefix and application name to provide default image
					createApplicationOptions.Image = fmt.Sprintf("%s/%s", prefix, fnName)
				} else if strings.HasPrefix(createApplicationOptions.Image, "_/") {
					// add the prefix to the specified image name
					createApplicationOptions.Image = strings.Replace(createApplicationOptions.Image, "_", prefix, 1)
				} else {
					return fmt.Errorf("Unknown image prefix syntax, expected %q, found: %q", "_/", createApplicationOptions.Image)
				}
			}

			return nil
		},
		Args: ArgValidationConjunction(
			cobra.ExactArgs(applicationCreateNumberOfArgs),
			AtPosition(applicationCreateApplicationNameIndex, ValidName()),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			fnName := args[applicationCreateApplicationNameIndex]

			createApplicationOptions.Name = fnName
			f, err := (*fcTool).CreateApplication(buildpackBuilder, createApplicationOptions, cmd.OutOrStdout())
			if err != nil {
				return err
			}

			if createApplicationOptions.DryRun {
				marshaller := NewMarshaller(cmd.OutOrStdout())
				if err = marshaller.Marshal(f); err != nil {
					return err
				}
			} else {
				PrintSuccessfulCompletion(cmd)
				if !createApplicationOptions.Verbose && !createApplicationOptions.Wait {
					namespaceOption := ""
					if createApplicationOptions.Namespace != "" {
						namespaceOption = fmt.Sprintf(" -n %s", createApplicationOptions.Namespace)
					}
					_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Issue `%s service status %s%s` to see the status of the application\n", env.Cli.Name, fnName, namespaceOption)
				}
			}

			return nil
		},
	}

	LabelArgs(command, "FUNCTION_NAME")

	command.Flags().StringVarP(&createApplicationOptions.Namespace, "namespace", "n", "", "the `namespace` of the service")
	command.Flags().BoolVarP(&createApplicationOptions.DryRun, "dry-run", "", false, dryRunUsage)
	command.Flags().StringVar(&createApplicationOptions.Image, "image", "", "the name of the image to build; must be a writable `repository/image[:tag]` with credentials configured")
	command.Flags().StringVar(&createApplicationOptions.GitRepo, "git-repo", "", "the `URL` for a git repository hosting the application code")
	command.Flags().StringVar(&createApplicationOptions.GitRevision, "git-revision", "master", "the git `ref-spec` of the application code to use")
	command.Flags().StringVar(&createApplicationOptions.SubPath, "sub-path", "", "the directory within the git repo to expose, files outside of this directory will not be available during the build")
	command.Flags().StringVarP(&createApplicationOptions.LocalPath, "local-path", "l", "", "`path` to local source to build the image from; only build-pack builds are supported at this time")
	command.Flags().BoolVarP(&createApplicationOptions.Verbose, "verbose", "v", false, verboseUsage)
	command.Flags().BoolVarP(&createApplicationOptions.Wait, "wait", "w", false, waitUsage)

	command.Flags().StringArrayVar(&createApplicationOptions.Env, "env", []string{}, envUsage)
	command.Flags().StringArrayVar(&createApplicationOptions.EnvFrom, "env-from", []string{}, envFromUsage)
	command.Flags().StringVar(&createApplicationOptions.BuildTemplate, "build-template", "cnb", "build template to apply")
	command.Flags().StringArrayVarP(&createApplicationOptions.Arguments, "argument", "a", []string{}, "build template arguments in a NAME=value form. Valid arguments will vary based on the build template")

	return command
}

func ApplicationUpdate(buildpackBuilder core.Builder, fcTool *core.Client) *cobra.Command {

	updateApplicationOptions := core.UpdateApplicationOptions{}

	command := &cobra.Command{
		Use:     "update",
		Short:   "Trigger a build to generate a new revision for a application resource",
		Example: `  ` + env.Cli.Name + ` application update square`,
		Args: ArgValidationConjunction(
			cobra.ExactArgs(applicationUpdateNumberOfArgs),
			AtPosition(applicationUpdateApplicationNameIndex, ValidName()),
		),
		RunE: func(cmd *cobra.Command, args []string) error {

			fnName := args[applicationUpdateApplicationNameIndex]

			updateApplicationOptions.Name = fnName
			err := (*fcTool).UpdateApplication(buildpackBuilder, updateApplicationOptions, cmd.OutOrStdout())
			if err != nil {
				return err
			}

			PrintSuccessfulCompletion(cmd)

			return nil
		},
	}

	LabelArgs(command, "APPLICATION_NAME")

	command.Flags().StringVarP(&updateApplicationOptions.Namespace, "namespace", "n", "", "the `namespace` of the application")
	command.Flags().StringVarP(&updateApplicationOptions.LocalPath, "local-path", "l", "", "path to local source to build the image from; only build-pack builds are supported at this time")
	command.Flags().BoolVarP(&updateApplicationOptions.Verbose, "verbose", "v", false, verboseUsage)
	command.Flags().BoolVarP(&updateApplicationOptions.Wait, "wait", "w", false, waitUsage)

	return command
}

func ApplicationBuild(buildpackBuilder core.Builder, fcTool *core.Client) *cobra.Command {
	buildApplicationOptions := core.BuildApplicationOptions{}

	command := &cobra.Command{
		Use:   "build",
		Short: "Build a application container from local source",
		Args:  cobra.ExactArgs(applicationBuildNumberOfArgs),
		RunE: func(cmd *cobra.Command, args []string) error {
			err := (*fcTool).BuildApplication(buildpackBuilder, buildApplicationOptions, cmd.OutOrStdout())
			if err != nil {
				return err
			}

			PrintSuccessfulCompletion(cmd)

			return nil
		},
	}

	command.Flags().StringVar(&buildApplicationOptions.Image, "image", "", "the name of the image to build; must be a writable `repository/image[:tag]` with credentials configured")
	command.MarkFlagRequired("image")
	command.Flags().StringVarP(&buildApplicationOptions.LocalPath, "local-path", "l", "", "`path` to local source to build the image from; only build-pack builds are supported at this time")
	command.MarkFlagRequired("local-path")

	return command
}

func ApplicationStatus(fcClient *core.Client) *cobra.Command {

	applicationStatusOptions := core.ApplicationStatusOptions{}

	command := &cobra.Command{
		Use:     "status",
		Short:   "Display the status of a application",
		Example: `  ` + env.Cli.Name + ` application status square --namespace joseph-ns`,
		Args: ArgValidationConjunction(
			cobra.ExactArgs(applicationStatusNumberOfArgs),
			AtPosition(applicationStatusApplicationNameIndex, ValidName()),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			fnName := args[applicationStatusApplicationNameIndex]
			applicationStatusOptions.Name = fnName
			cond, err := (*fcClient).ApplicationStatus(applicationStatusOptions)
			if err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Last Transition Time:        %s\n", cond.LastTransitionTime.Inner.Format(time.RFC3339))

			if cond.Reason != "" {
				fmt.Fprintf(cmd.OutOrStdout(), "Message:                     %s\n", cond.Message)
				fmt.Fprintf(cmd.OutOrStdout(), "Reason:                      %s\n", cond.Reason)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Status:                      %s\n", cond.Status)
			fmt.Fprintf(cmd.OutOrStdout(), "Type:                        %s\n", cond.Type)

			return nil
		},
	}

	LabelArgs(command, "FUNCTION_NAME")

	command.Flags().StringVarP(&applicationStatusOptions.Namespace, "namespace", "n", "", "the `namespace` of the application")

	return command
}

func ApplicationList(fcClient *core.Client) *cobra.Command {
	listApplicationOptions := core.ListApplicationOptions{}

	command := &cobra.Command{
		Use:   "list",
		Short: "List application resources",
		Example: `  ` + env.Cli.Name + ` application list
  ` + env.Cli.Name + ` application list --namespace joseph-ns`,
		Args: cobra.ExactArgs(applicationListNumberOfArgs),
		RunE: func(cmd *cobra.Command, args []string) error {
			applications, err := (*fcClient).ListApplications(listApplicationOptions)
			if err != nil {
				return err
			}

			Display(cmd.OutOrStdout(), applicationToInterfaceSlice(applications.Items), makeApplicationExtractors())
			PrintSuccessfulCompletion(cmd)
			return nil
		},
	}

	command.Flags().StringVarP(&listApplicationOptions.Namespace, "namespace", "n", "", "the `namespace` of the applications to be listed")

	return command
}

func ApplicationInvoke(fcClient *core.Client) *cobra.Command {

	applicationInvokeOptions := core.ApplicationInvokeOptions{}

	command := &cobra.Command{
		Use:   "invoke",
		Short: "Invoke a application",
		Long: `Invoke a application by shelling out to curl.

The curl command is printed so it can be copied and extended.

Additional curl arguments and flags may be specified after a double dash (--).`,
		Example: `  ` + env.Cli.Name + ` application invoke square --namespace joseph-ns
  ` + env.Cli.Name + ` application invoke square /foo -- --data 42`,
		PreRunE: FlagsValidatorAsCobraRunE(AtMostOneOf("json", "text")),
		Args: UpToDashDash(ArgValidationConjunction(
			cobra.MinimumNArgs(applicationInvokeMaxNumberOfArgs-1),
			cobra.MaximumNArgs(applicationInvokeMaxNumberOfArgs),
			AtPosition(applicationInvokeServiceNameIndex, ValidName()),
		)),
		RunE: func(cmd *cobra.Command, args []string) error {
			argsLengthAtDash := cmd.ArgsLenAtDash()
			applicationInvokeOptions.Name = args[applicationInvokeServiceNameIndex]
			path := "/"
			if argsLengthAtDash > applicationInvokeServicePathIndex ||
				argsLengthAtDash == -1 && len(args) > applicationInvokeServicePathIndex {
				path = args[applicationInvokeServicePathIndex]
			}
			ingress, hostName, err := (*fcClient).ApplicationCoordinates(applicationInvokeOptions)
			if err != nil {
				return err
			}

			curlCmd := exec.Command("curl", ingress+path)

			curlCmd.Stdin = os.Stdin
			curlCmd.Stdout = cmd.OutOrStdout()
			curlCmd.Stderr = cmd.OutOrStderr()

			hostHeader := fmt.Sprintf("Host: %s", hostName)
			curlCmd.Args = append(curlCmd.Args, "-H", hostHeader)

			if applicationInvokeOptions.ContentTypeJson {
				curlCmd.Args = append(curlCmd.Args, "-H", "Content-Type: application/json")
			} else if applicationInvokeOptions.ContentTypeText {
				curlCmd.Args = append(curlCmd.Args, "-H", "Content-Type: text/plain")
			}

			verbose := false

			if argsLengthAtDash > 0 {
				curlArgs := args[argsLengthAtDash:]
				for _, a := range curlArgs {
					if verboseCurl(a) {
						verbose = true
					}
				}
				curlCmd.Args = append(curlCmd.Args, curlArgs...)
			}

			quoted, err := shellquote.Quote(curlCmd.Args)
			if err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), quoted)

			if verbose {
				return curlCmd.Run()
			}

			// curl is not verbose, so make it verbose, capture standard error, and print any HTTP errors
			curlCmd.Args = append(curlCmd.Args, "-v")

			buffer := new(bytes.Buffer)
			errStream := curlCmd.Stderr
			curlCmd.Stderr = buffer

			curlErr := curlCmd.Run()

			PrintCurlHttpErrors(buffer.String(), errStream)

			return curlErr
		},
	}

	LabelArgs(command, "FUNCTION_NAME", "PATH")

	command.Flags().StringVarP(&applicationInvokeOptions.Namespace, "namespace", "n", "", "the `namespace` of the service")
	command.Flags().BoolVar(&applicationInvokeOptions.ContentTypeJson, "json", false, "set the request's content type to 'application/json'")
	command.Flags().BoolVar(&applicationInvokeOptions.ContentTypeText, "text", false, "set the request's content type to 'text/plain'")

	return command
}

func ApplicationDelete(riffClient *core.Client) *cobra.Command {
	cliOptions := DeleteApplicationsCliOptions{}

	command := &cobra.Command{
		Use:   "delete",
		Short: "Delete existing applications",
		Example: `  ` + env.Cli.Name + ` application delete square --namespace joseph-ns
  ` + env.Cli.Name + ` application delete service-1 service-2`,
		Args: ArgValidationConjunction(
			cobra.MinimumNArgs(applicationDeleteMinNumberOfArgs),
			StartingAtPosition(applicationDeleteNameStartIndex, ValidName()),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			names := args[applicationDeleteNameStartIndex:]
			results := tasks.ApplyInParallel(names, func(name string) error {
				options := core.DeleteApplicationOptions{Namespace: cliOptions.Namespace, Name: name}
				return (*riffClient).DeleteApplication(options)
			})
			err := tasks.MergeResults(results, func(result tasks.CorrelatedResult) string {
				err := result.Error
				if err == nil {
					return ""
				}
				return fmt.Sprintf("Unable to delete application %s: %v", result.Input, err)
			})
			if err != nil {
				return err
			}

			PrintSuccessfulCompletion(cmd)
			return nil
		},
	}

	LabelArgs(command, "FUNCTION_NAME")

	command.Flags().StringVarP(&cliOptions.Namespace, "namespace", "n", "", "the `namespace` of the service")

	return command
}

func applicationToInterfaceSlice(applications []projectriffv1alpha1.Application) []interface{} {
	result := make([]interface{}, len(applications))
	for i := range applications {
		result[i] = applications[i]
	}
	return result
}

func makeApplicationExtractors() []NamedExtractor {
	return []NamedExtractor{
		{
			name: "NAME",
			fn:   func(f interface{}) string { return f.(projectriffv1alpha1.Application).Name },
		},
		{
			name: "STATUS",
			fn: func(f interface{}) string {
				application := f.(projectriffv1alpha1.Application)
				cond := application.Status.GetCondition(projectriffv1alpha1.ApplicationConditionReady)
				if cond == nil {
					return "Unknown"
				}
				switch cond.Status {
				case v1.ConditionTrue:
					return "Running"
				case v1.ConditionFalse:
					return fmt.Sprintf("%s: %s", cond.Reason, cond.Message)
				default:
					return "Unknown"
				}
			},
		},
	}
}

type DeleteApplicationsCliOptions struct {
	Namespace string
}
