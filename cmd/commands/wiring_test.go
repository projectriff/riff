package commands_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"github.com/projectriff/riff/cmd/commands"
	"github.com/spf13/cobra"
)

var _ = Describe("`riff` root command", func() {
	Context("include subcommands", func() {
		var rootCommand *cobra.Command

		BeforeEach(func() {
			rootCommand = commands.CreateAndWireRootCommand()
		})

		DescribeTable("which define their own",
			func(subcommandName string, subsubcommandName string) {
				serviceCmd := matchSubcommandByName(rootCommand, subcommandName)

				Expect(serviceCmd).NotTo(BeNil(), "root command should include subcommand " + subcommandName)
				Expect(commandNamesOf(serviceCmd.Commands())).To(ContainElement(subsubcommandName))
			},
			Entry("∃ `subscription create`", "subscription", "create"),
			Entry("∃ `subscription delete`", "subscription", "delete"),
			Entry("∃ `subscription list`", "subscription", "list"),
		)
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
