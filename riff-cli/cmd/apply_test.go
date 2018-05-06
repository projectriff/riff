package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/projectriff/riff/riff-cli/pkg/kubectl"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/mock"
)

var _ = Describe("The apply command", func() {

	var (
		oldCWD        string
		realKubeCtl   *kubectl.MockKubeCtl
		dryRunKubeCtl *kubectl.MockKubeCtl
		applyCmd      *cobra.Command
		args          []string
	)

	BeforeEach(func() {
		var err error

		oldCWD, err = os.Getwd()
		Expect(err).NotTo(HaveOccurred())

		realKubeCtl = new(kubectl.MockKubeCtl)
		dryRunKubeCtl = new(kubectl.MockKubeCtl)

		applyCmd, _ = Apply(realKubeCtl, dryRunKubeCtl)
		args = []string{}

	})

	AfterEach(func() {
		os.Chdir(oldCWD)
	})

	Context("with no --filepath provided", func() {
		BeforeEach(func() {
			os.Chdir("../test_data/command/fn-with-existing-files")
		})

		It("should apply from the current directory", func() {

			topicsFile, err := filepath.Abs("echo-topics.yaml")
			Expect(err).NotTo(HaveOccurred())
			applesFunctionFile, err := filepath.Abs("apples-function.yaml")
			Expect(err).NotTo(HaveOccurred())
			applesTopicBindingFile, err := filepath.Abs("apples-topicbinding.yaml")
			Expect(err).NotTo(HaveOccurred())
			orangesFunctionFile, err := filepath.Abs("oranges-function.yaml")
			Expect(err).NotTo(HaveOccurred())
			orangesTopicBindingFile, err := filepath.Abs("oranges-topicbinding.yaml")
			Expect(err).NotTo(HaveOccurred())

			realKubeCtl.On("Exec", []string{
				"apply",
				"-f", applesFunctionFile,
				"-f", orangesFunctionFile,
				"-f", topicsFile,
				"-f", applesTopicBindingFile,
				"-f", orangesTopicBindingFile,
			}).Return("", nil)

			err = applyCmd.Execute()
			Expect(err).NotTo(HaveOccurred())

		})

		It("should use the provided namespace if present", func() {

			topicsFile, err := filepath.Abs("echo-topics.yaml")
			Expect(err).NotTo(HaveOccurred())
			applesFunctionFile, err := filepath.Abs("apples-function.yaml")
			Expect(err).NotTo(HaveOccurred())
			applesTopicBindingFile, err := filepath.Abs("apples-topicbinding.yaml")
			Expect(err).NotTo(HaveOccurred())
			orangesFunctionFile, err := filepath.Abs("oranges-function.yaml")
			Expect(err).NotTo(HaveOccurred())
			orangesTopicBindingFile, err := filepath.Abs("oranges-topicbinding.yaml")
			Expect(err).NotTo(HaveOccurred())

			realKubeCtl.On("Exec", []string{
				"apply",
				"--namespace", "foobar",
				"-f", applesFunctionFile,
				"-f", orangesFunctionFile,
				"-f", topicsFile,
				"-f", applesTopicBindingFile,
				"-f", orangesTopicBindingFile,
			}).Return("", nil)

			applyCmd.SetArgs([]string{"--namespace", "foobar"})
			err = applyCmd.Execute()
			Expect(err).NotTo(HaveOccurred())

		})

		It("should not do anything for real if dry-run is set", func() {

			topicsFile, err := filepath.Abs("echo-topics.yaml")
			Expect(err).NotTo(HaveOccurred())
			applesFunctionFile, err := filepath.Abs("apples-function.yaml")
			Expect(err).NotTo(HaveOccurred())
			applesTopicBindingFile, err := filepath.Abs("apples-topicbinding.yaml")
			Expect(err).NotTo(HaveOccurred())
			orangesFunctionFile, err := filepath.Abs("oranges-function.yaml")
			Expect(err).NotTo(HaveOccurred())
			orangesTopicBindingFile, err := filepath.Abs("oranges-topicbinding.yaml")
			Expect(err).NotTo(HaveOccurred())

			dryRunKubeCtl.On("Exec", []string{
				"apply",
				"-f", applesFunctionFile,
				"-f", orangesFunctionFile,
				"-f", topicsFile,
				"-f", applesTopicBindingFile,
				"-f", orangesTopicBindingFile,
			}).Return("", nil)

			applyCmd.SetArgs([]string{"--dry-run"})
			err = applyCmd.Execute()
			Expect(err).NotTo(HaveOccurred())

		})
	})

	It("should accept a directory as an arg", func() {
		topicsFile, err := filepath.Abs("../test_data/command/fn-with-existing-files/echo-topics.yaml")
		Expect(err).NotTo(HaveOccurred())
		applesFunctionFile, err := filepath.Abs("../test_data/command/fn-with-existing-files/apples-function.yaml")
		Expect(err).NotTo(HaveOccurred())
		applesTopicBindingFile, err := filepath.Abs("../test_data/command/fn-with-existing-files/apples-topicbinding.yaml")
		Expect(err).NotTo(HaveOccurred())
		orangesFunctionFile, err := filepath.Abs("../test_data/command/fn-with-existing-files/oranges-function.yaml")
		Expect(err).NotTo(HaveOccurred())
		orangesTopicBindingFile, err := filepath.Abs("../test_data/command/fn-with-existing-files/oranges-topicbinding.yaml")
		Expect(err).NotTo(HaveOccurred())

		realKubeCtl.On("Exec", []string{
			"apply",
			"-f", applesFunctionFile,
			"-f", orangesFunctionFile,
			"-f", topicsFile,
			"-f", applesTopicBindingFile,
			"-f", orangesTopicBindingFile,
		}).Return("", nil)

		applyCmd.SetArgs([]string{"../test_data/command/fn-with-existing-files"})
		err = applyCmd.Execute()
		Expect(err).NotTo(HaveOccurred())

	})

	It("should accept a directory as a flag", func() {
		topicsFile, err := filepath.Abs("../test_data/command/fn-with-existing-files/echo-topics.yaml")
		Expect(err).NotTo(HaveOccurred())
		applesFunctionFile, err := filepath.Abs("../test_data/command/fn-with-existing-files/apples-function.yaml")
		Expect(err).NotTo(HaveOccurred())
		applesTopicBindingFile, err := filepath.Abs("../test_data/command/fn-with-existing-files/apples-topicbinding.yaml")
		Expect(err).NotTo(HaveOccurred())
		orangesFunctionFile, err := filepath.Abs("../test_data/command/fn-with-existing-files/oranges-function.yaml")
		Expect(err).NotTo(HaveOccurred())
		orangesTopicBindingFile, err := filepath.Abs("../test_data/command/fn-with-existing-files/oranges-topicbinding.yaml")
		Expect(err).NotTo(HaveOccurred())

		realKubeCtl.On("Exec", []string{
			"apply",
			"-f", applesFunctionFile,
			"-f", orangesFunctionFile,
			"-f", topicsFile,
			"-f", applesTopicBindingFile,
			"-f", orangesTopicBindingFile,
		}).Return("", nil)

		applyCmd.SetArgs([]string{"--filepath", "../test_data/command/fn-with-existing-files"})
		err = applyCmd.Execute()
		Expect(err).NotTo(HaveOccurred())

	})

	It("should accept a single file as a flag", func() {
		applesFile, err := filepath.Abs("../test_data/command/fn-with-existing-files/apples-function.yaml")
		Expect(err).NotTo(HaveOccurred())

		realKubeCtl.On("Exec", []string{"apply", "-f", applesFile}).Return("", nil)

		applyCmd.SetArgs([]string{"--filepath", "../test_data/command/fn-with-existing-files/apples-function.yaml"})
		err = applyCmd.Execute()
		Expect(err).NotTo(HaveOccurred())

	})

	It("should accept a single file as an arg", func() {
		applesFile, err := filepath.Abs("../test_data/command/fn-with-existing-files/apples-function.yaml")
		Expect(err).NotTo(HaveOccurred())

		realKubeCtl.On("Exec", []string{"apply", "-f", applesFile}).Return("", nil)

		applyCmd.SetArgs([]string{"../test_data/command/fn-with-existing-files/apples-function.yaml"})
		err = applyCmd.Execute()
		Expect(err).NotTo(HaveOccurred())

	})

	It("should report kubectl errors", func() {
		applyCmd.SetArgs([]string{"../test_data/command/fn-with-existing-files"})
		realKubeCtl.On("Exec", mock.MatchedBy(func(interface{}) bool { return true })).Return("", fmt.Errorf("Whoops"))

		err := applyCmd.Execute()
		Expect(err).To(MatchError("Whoops"))

	})

})
