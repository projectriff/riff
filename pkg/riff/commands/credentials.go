package commands

import (
	"github.com/projectriff/riff/pkg/core"
	"github.com/projectriff/riff/pkg/env"
	"github.com/spf13/cobra"
)

const (
	secretNameIndex = iota
	credentialsSetNumberOfArgs
)

func Credentials() *cobra.Command {
	return &cobra.Command{
		Use:   "credentials",
		Short: "Interact with credentials related resources",
	}
}

func CredentialsSet(c *core.Client) *cobra.Command {
	options := core.SetCredentialsOptions{}

	command := &cobra.Command{
		Use:     "set",
		Short:   "create or update secret and bind it to the " + env.Cli.Name + " service account",
		Example: `  ` + env.Cli.Name + ` credentials set build-secret --namespace default --docker-hub johndoe`,
		Args: ArgValidationConjunction(
			cobra.ExactArgs(credentialsSetNumberOfArgs),
			AtPosition(secretNameIndex, ArgNotBlank()),
		),
		PreRunE: FlagsValidatorAsCobraRunE(
			FlagsValidationConjunction(
				FlagsDependency(Set("namespace"), ValidDnsSubdomain("namespace")),
				AtMostOneOf("gcr", "docker-hub", "registry-user"),
				FlagsDependency(Set("registry-user"), NotBlank("registry")),
				FlagsDependency(Set("registry"),
					NotBlank("registry-user"),
					SupportedRegistryProtocol(func() string {
						return options.Registry
					}))),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			options.SecretName = args[secretNameIndex]
			if err := (*c).SetCredentials(options); err != nil {
				return err
			}
			PrintSuccessfulCompletion(cmd)
			return nil
		},
	}

	command.Flags().StringVar(&options.NamespaceName, "namespace", "", "the `namespace` of the credentials to be added")
	command.Flags().StringVarP(&options.SecretName, "secret", "s", "", "the name of a `secret` containing credentials for the image registry")
	command.Flags().StringVar(&options.GcrTokenPath, "gcr", "", "path to a file containing Google Container Registry credentials")
	command.Flags().StringVar(&options.DockerHubId, "docker-hub", "", "Docker ID for authenticating with Docker Hub; password will be read from stdin")
	command.Flags().StringVar(&options.Registry, "registry", "", "registry server host, scheme must be \"http\" or \"https\" (default \"https\")")
	command.Flags().StringVar(&options.RegistryUser, "registry-user", "", "registry username; password will be read from stdin")

	return command
}
