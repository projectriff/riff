package commands_test

import (
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/projectriff/riff/cmd/commands"
	"github.com/spf13/cobra"
)

var _ = Describe("`riff` root command", func() {
	Context("should wire subcommands", func() {
		var rootCommand *cobra.Command

		BeforeEach(func() {
			rootCommand = commands.CreateAndWireRootCommand()
		})

		It("including `riff subscription`", func() {
			errMsg := "`%s` should be wired to root command"
			Expect(find(rootCommand, "subscription")).NotTo(BeNil(), fmt.Sprintf(errMsg, "subscription"))
			Expect(find(rootCommand, "subscription", "create")).NotTo(BeNil(), fmt.Sprintf(errMsg, "subscription create"))
			Expect(find(rootCommand, "subscription", "delete")).NotTo(BeNil(), fmt.Sprintf(errMsg, "subscription delete"))
			Expect(find(rootCommand, "subscription", "list")).NotTo(BeNil(), fmt.Sprintf(errMsg, "subscription list"))
		})

		It("with proper kubeclient init", func() {
			subscription := find(rootCommand, "subscription")
			Expect(subscription.Flag("kubeconfig")).NotTo(BeNil())
			Expect(subscription.Flag("master")).NotTo(BeNil())
		})

	})

})

func find(command *cobra.Command, names ...string) *cobra.Command {
	cmd, unmatchedArgs, err := command.Find(names)
	if err != nil || len(unmatchedArgs) > 0 {
		return nil
	}
	return cmd
}
