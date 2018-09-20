package commands_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/projectriff/riff/cmd/commands"
	"github.com/spf13/cobra"
)

var _ = Describe("`riff` root command", func() {
	Context("wire subcommands", func() {
		var rootCommand *cobra.Command

		BeforeEach(func() {
			rootCommand = commands.CreateAndWireRootCommand()
		})

		Context("such as `riff` subscription, which should", func() {
			var (
				subscriptionCmd *cobra.Command
			)

			BeforeEach(func() {
				subscriptionCmd = matchSubcommandByName(rootCommand, "subscription")
				Expect(subscriptionCmd).NotTo(BeNil(), "`subscription` should be wired to root command")
			})

			It("wire `subscription create`", func() {
				Expect(commandNamesOf(subscriptionCmd.Commands())).To(ContainElement("create"))
			})

			It("wire `subscription delete`", func() {
				Expect(commandNamesOf(subscriptionCmd.Commands())).To(ContainElement("delete"))
			})

			It("wire `subscription list`", func() {
				Expect(commandNamesOf(subscriptionCmd.Commands())).To(ContainElement("list"))
			})
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
