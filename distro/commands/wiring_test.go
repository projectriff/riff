package commands_test

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/projectriff/riff/cmd/commands"
	"github.com/projectriff/riff/distro/commands"
	"github.com/projectriff/riff/pkg/core"
	"github.com/spf13/cobra"
)

var _ = Describe("`riff-distro` root command", func() {
	Context("should wire subcommands", func() {
		var rootCommand *cobra.Command
		var manifests map[string]*core.Manifest

		BeforeEach(func() {
			rootCommand = commands.DistroCreateAndWireRootCommand(manifests)
		})

		It("including `riff-distro` docs", func() {
			errMsg := "`%s` should be wired to root command"
			Expect(FindSubcommand(rootCommand, "docs")).NotTo(BeNil(), fmt.Sprintf(errMsg, "docs"))
		})

	})

})
