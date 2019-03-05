package fileutils_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/projectriff/riff/pkg/fileutils"
)

var _ = Describe("StartsWithHomeDirAsTilde", func() {

	It("returns true when starting with ~/", func() {
		result := fileutils.StartsWithCurrentUserDirectoryAsTilde("~/", "darwin")

		Expect(result).To(BeTrue(), "tilde+forward slash should work")
	})

	It(`returns false when starting with ~\ on Mac OS`, func() {
		result := fileutils.StartsWithCurrentUserDirectoryAsTilde(`~\`, "darwin")

		Expect(result).To(BeFalse(), "tilde+backslash should not work on Mac OS")
	})

	It(`returns true when starting with ~\ on Windows`, func() {
		result := fileutils.StartsWithCurrentUserDirectoryAsTilde(`~\`, "windows")

		Expect(result).To(BeTrue(), "tilde+backslash on Windows should work")
	})

	It(`returns true when starting with ~/ on Windows`, func() {
		result := fileutils.StartsWithCurrentUserDirectoryAsTilde(`~/`, "windows")

		Expect(result).To(BeTrue(), "tilde+forward slash on Windows should work")
	})
})

var _ = Describe("ResolveTilde", func() {

	It("resolves ~/ against current user's home directory", func() {
		initialPath := "~/some/location"

		path, err := fileutils.ResolveTilde(initialPath)

		Expect(err).NotTo(HaveOccurred())
		Expect(path).NotTo(ContainSubstring("~"))
		Expect(path).To(HaveSuffix(initialPath[2:]))
	})

	It("returns path without tilde as is", func() {
		initialPath := "look/matilde/no/tilde"

		path, err := fileutils.ResolveTilde(initialPath)

		Expect(err).NotTo(HaveOccurred())
		Expect(path).To(Equal(initialPath))
	})

	It("returns path with embedded tilde as is", func() {
		initialPath := "look/matilde/thereisa/~"

		path, err := fileutils.ResolveTilde(initialPath)

		Expect(err).NotTo(HaveOccurred())
		Expect(path).To(Equal(initialPath))
	})
})
