package commands_test

import (
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/projectriff/riff/cmd/commands"
	"github.com/projectriff/riff/distro/commands"
	"github.com/spf13/cobra"
)

var _ = Describe("`riff-distro` root command", func() {
	Context("should wire subcommands", func() {
		var rootCommand *cobra.Command

		BeforeEach(func() {
			rootCommand = commands.DistroCreateAndWireRootCommand()
		})

		It("including `riff-distro` docs", func() {
			errMsg := "`%s` should be wired to root command"
			Expect(FindSubcommand(rootCommand, "docs")).NotTo(BeNil(), fmt.Sprintf(errMsg, "docs"))
		})

	})

})
