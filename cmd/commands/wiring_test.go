package commands_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/projectriff/riff/cmd/commands"
	"github.com/spf13/cobra"
)

var _ = Describe("`riff` root command", func() {
	var rootCommand *cobra.Command

	BeforeEach(func() {
		rootCommand = commands.CreateAndWireRootCommand()
	})

	It("should include the `service` subcommand", func() {
		Expect(namesOf(rootCommand.Commands())).To(ContainElement("service"))
	})

	It("should include `unsubscribe` as a `service` subcommand", func() {
		serviceCmd := matchSubcommandByName(rootCommand, "service")

		Expect(namesOf(serviceCmd.Commands())).To(ContainElement("unsubscribe"))
	})
})

func namesOf(commands []*cobra.Command) []string {
	result := make([]string, len(commands))
	for _, e := range commands {
		result = append(result, e.Name())
	}
	return result
}

func matchSubcommandByName(command *cobra.Command, name string) *cobra.Command {
	for _, e := range command.Commands() {
		if e.Name() == name {
			return e
		}
	}
	return nil
}