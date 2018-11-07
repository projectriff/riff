package commands_test

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/projectriff/riff/cmd/commands"
	"github.com/projectriff/riff/pkg/core"
	"github.com/spf13/cobra"
)

var _ = Describe("`riff` root command", func() {
	Context("should wire subcommands", func() {
		var (
			rootCommand *cobra.Command
			manifests   map[string]*core.Manifest
		)

		BeforeEach(func() {
			rootCommand = CreateAndWireRootCommand(manifests, "")
		})

		It("including `riff subscription`", func() {
			errMsg := "`%s` should be wired to root command"
			Expect(FindSubcommand(rootCommand, "subscription")).NotTo(BeNil(), fmt.Sprintf(errMsg, "subscription"))
			Expect(FindSubcommand(rootCommand, "subscription", "create")).NotTo(BeNil(), fmt.Sprintf(errMsg, "subscription create"))
			Expect(FindSubcommand(rootCommand, "subscription", "delete")).NotTo(BeNil(), fmt.Sprintf(errMsg, "subscription delete"))
			Expect(FindSubcommand(rootCommand, "subscription", "list")).NotTo(BeNil(), fmt.Sprintf(errMsg, "subscription list"))
		})

	})

})
