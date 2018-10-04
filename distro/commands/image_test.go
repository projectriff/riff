package commands_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/projectriff/riff/distro/commands"
	"github.com/projectriff/riff/pkg/core"
	"github.com/projectriff/riff/pkg/core/mocks"
	"github.com/spf13/cobra"
)

var _ = Describe("The riff-distro image list command", func() {

	Context("when given suitable flags", func() {
		var (
			client          core.ImageClient
			clientMock      *mocks.ImageClient
			command         *cobra.Command
			expectedOptions core.ListImagesOptions
		)

		BeforeEach(func() {
			client = new(mocks.ImageClient)
			clientMock = client.(*mocks.ImageClient)
			command = commands.ImageList(&client)
		})

		AfterEach(func() {
			clientMock.AssertExpectations(GinkgoT())
		})

		It("should have sensible defaults", func() {
			expectedOptions = core.ListImagesOptions{
				Manifest: "stable",
				Images:   "",
				NoCheck:  false,
				Force:    false,
			}
			clientMock.On("ListImages", expectedOptions).Return(nil)

			command.SetArgs([]string{})
			err := command.Execute()

			Expect(err).NotTo(HaveOccurred())
		})

		It("should not check images when the corresponding flag is explicitly set", func() {
			expectedOptions = core.ListImagesOptions{
				Manifest: "stable",
				Images:   "",
				NoCheck:  true,
				Force:    false,
			}
			clientMock.On("ListImages", expectedOptions).Return(nil)

			command.SetArgs([]string{"--no-check=true"})
			err := command.Execute()

			Expect(err).NotTo(HaveOccurred())
		})

		It("should not check images when the corresponding flag is set without value", func() {
			expectedOptions = core.ListImagesOptions{
				Manifest: "stable",
				Images:   "",
				NoCheck:  true,
				Force:    false,
			}
			clientMock.On("ListImages", expectedOptions).Return(nil)

			command.SetArgs([]string{"--no-check"})
			err := command.Execute()

			Expect(err).NotTo(HaveOccurred())
		})
	})
})
