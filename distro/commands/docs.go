package commands

import (
	"github.com/projectriff/riff/cmd/commands"
	"github.com/spf13/cobra"
)

func DistroDocs(rootCommand *cobra.Command, fs commands.Filesystem) *cobra.Command {
	var directory string

	command := &cobra.Command{
		Use:    "docs",
		Short:  "generate riff-distro command documentation",
		Hidden: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return commands.GenerateDocs(rootCommand, directory, fs)
		},
	}

	command.Flags().StringVarP(&directory, "dir", "d", "docs", "the output directory for the docs.")
	return command
}
