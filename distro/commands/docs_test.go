package commands

import (
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/projectriff/riff/cmd/commands"
	"github.com/spf13/cobra"
	"io/ioutil"
	"os"
	"path/filepath"
)

type mkdirKoFs struct{}

func (mkdirKoFs) Mkdir(name string, perm os.FileMode) error {
	return fmt.Errorf("oopsie, Mkdir failed")
}
func (mkdirKoFs) Stat(name string) (os.FileInfo, error) {
	return os.Stat(name)
}

type statKoFs struct{}

func (statKoFs) Mkdir(name string, perm os.FileMode) error {
	return os.Mkdir(name, perm)
}
func (statKoFs) Stat(name string) (os.FileInfo, error) {
	return nil, fmt.Errorf("oopsie, Stat failed")
}

var _ = Describe("The riff-distro docs command", func() {
	var (
		command *cobra.Command
	)

	Context("when given incorrect args or flags", func() {

		It("should fail when specified path is not a directory", func() {
			command = DistroDocs(nil, commands.LocalFs{})
			notADir, err := ioutil.TempFile("", "invalid-distro-docs")
			Expect(err).To(BeNil(), "could not create file")
			defer notADir.Close()
			defer os.Remove(notADir.Name())
			command.SetArgs([]string{"--dir", notADir.Name()})

			err = command.Execute()

			Expect(err).To(MatchError(ContainSubstring("already exists but is not a directory")))
		})

		It("should fail when a directory cannot be created from specified path", func() {
			command = DistroDocs(nil, mkdirKoFs{})
			command.SetArgs([]string{"--dir", "mkdir-will-fail"})

			err := command.Execute()

			Expect(err).To(MatchError("oopsie, Mkdir failed"))
		})

		It("should fail when getting file information fails", func() {
			command = DistroDocs(nil, statKoFs{})
			dir, err := ioutil.TempDir("", "distro-docs-ko-stat")
			Expect(err).To(BeNil(), "could not create dir")
			defer os.Remove(dir)
			command.SetArgs([]string{"--dir", dir})

			err = command.Execute()

			Expect(err).To(MatchError("oopsie, Stat failed"))
		})
	})

	Context("when given suitable args and flags", func() {

		BeforeEach(func() {
			command = DistroDocs(DistroCreateAndWireRootCommand(), commands.LocalFs{})
		})

		It("is properly documented", func() {
			Expect(command.Name()).To(Equal("docs"))
			Expect(command.Short).To(Not(BeEmpty()))
		})

		It("is hidden", func() {
			Expect(command.Hidden).To(BeTrue())
		})

		It("generates docs in the specified directory", func() {
			dir, err := ioutil.TempDir("", "distro-docs")
			Expect(err).To(BeNil(), "could not create dir")
			defer os.Remove(dir)
			command.SetArgs([]string{"--dir", dir})

			err = command.Execute()

			Expect(err).To(BeNil())
			files, err := matchFiles(dir + "/*.md")
			for _, file := range files {
				defer os.Remove(file)
			}
			Expect(err).Should(BeNil())
			Expect(len(files)).Should(BeNumerically(">", 0))
		})
	})
})

func matchFiles(pattern string) ([]string, error) {
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, err
	}
	return matches, nil
}
