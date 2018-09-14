package commands_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/projectriff/riff/cmd/commands"
	"github.com/spf13/cobra"
)

var _ = Describe("`riff` root command", func() {
	Context("subscription", func() {
		var rootCommand *cobra.Command

		BeforeEach(func() {
			rootCommand = commands.CreateAndWireRootCommand()
		})

		It("should be included in riff subcommands", func() {
			Expect(commandNamesOf(rootCommand.Commands())).To(ContainElement("subscription"))
		})

		It("should define a `create` subcommand", func() {
			serviceCmd := matchSubcommandByName(rootCommand, "subscription")

			Expect(commandNamesOf(serviceCmd.Commands())).To(ContainElement("create"))
		})

		It("should define a `delete` subcommand", func() {
			serviceCmd := matchSubcommandByName(rootCommand, "subscription")

			Expect(commandNamesOf(serviceCmd.Commands())).To(ContainElement("delete"))
		})
	})

})

func commandNamesOf(commands []*cobra.Command) []string {
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
