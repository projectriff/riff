package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/projectriff/riff/riff-cli/pkg/kubectl"
	"github.com/spf13/cobra"
)

var _ = Describe("The apply command", func() {

	const canned_kubectl_get_response = `{
				"apiVersion": "projectriff.io/v1alpha1",
				"kind": "Function",
				"metadata": {},
				"spec": {
					"container": {
					"image": "test/echo:0.0.1"
					},
					"input": "myInputTopic",
					"output": "myOutputTopic",
					"protocol": "grpc"
				}
			}`

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
			applesFile, err := filepath.Abs("apples-function.yaml")
			Expect(err).NotTo(HaveOccurred())
			orangesFile, err := filepath.Abs("oranges-function.yaml")
			Expect(err).NotTo(HaveOccurred())

			realKubeCtl.On("Exec", []string{"apply", "-f", applesFile, "-f", orangesFile, "-f", topicsFile}).Return("", nil)

			err = applyCmd.Execute()
			Expect(err).NotTo(HaveOccurred())

		})

		It("should use the provided namespace if present", func() {

			topicsFile, err := filepath.Abs("echo-topics.yaml")
			Expect(err).NotTo(HaveOccurred())
			applesFile, err := filepath.Abs("apples-function.yaml")
			Expect(err).NotTo(HaveOccurred())
			orangesFile, err := filepath.Abs("oranges-function.yaml")
			Expect(err).NotTo(HaveOccurred())

			realKubeCtl.On("Exec", []string{"apply", "--namespace", "foobar", "-f", applesFile, "-f", orangesFile, "-f", topicsFile}).Return("", nil)

			applyCmd.SetArgs([]string{"--namespace", "foobar"})
			err = applyCmd.Execute()
			Expect(err).NotTo(HaveOccurred())

		})

		It("should not do anything for real if dry-run is set", func() {

			topicsFile, err := filepath.Abs("echo-topics.yaml")
			Expect(err).NotTo(HaveOccurred())
			applesFile, err := filepath.Abs("apples-function.yaml")
			Expect(err).NotTo(HaveOccurred())
			orangesFile, err := filepath.Abs("oranges-function.yaml")
			Expect(err).NotTo(HaveOccurred())

			dryRunKubeCtl.On("Exec", []string{"apply", "-f", applesFile, "-f", orangesFile, "-f", topicsFile}).Return("", nil)

			applyCmd.SetArgs([]string{"--dry-run"})
			err = applyCmd.Execute()
			Expect(err).NotTo(HaveOccurred())

		})
	})

	It("should accept a directory as an arg", func() {
		topicsFile, err := filepath.Abs("../test_data/command/fn-with-existing-files/echo-topics.yaml")
		Expect(err).NotTo(HaveOccurred())
		applesFile, err := filepath.Abs("../test_data/command/fn-with-existing-files/apples-function.yaml")
		Expect(err).NotTo(HaveOccurred())
		orangesFile, err := filepath.Abs("../test_data/command/fn-with-existing-files/oranges-function.yaml")
		Expect(err).NotTo(HaveOccurred())

		realKubeCtl.On("Exec", []string{"apply", "-f", applesFile, "-f", orangesFile, "-f", topicsFile}).Return("", nil)

		applyCmd.SetArgs([]string{"../test_data/command/fn-with-existing-files"})
		err = applyCmd.Execute()
		Expect(err).NotTo(HaveOccurred())

	})

	It("should accept a directory as a flag", func() {
		topicsFile, err := filepath.Abs("../test_data/command/fn-with-existing-files/echo-topics.yaml")
		Expect(err).NotTo(HaveOccurred())
		applesFile, err := filepath.Abs("../test_data/command/fn-with-existing-files/apples-function.yaml")
		Expect(err).NotTo(HaveOccurred())
		orangesFile, err := filepath.Abs("../test_data/command/fn-with-existing-files/oranges-function.yaml")
		Expect(err).NotTo(HaveOccurred())

		realKubeCtl.On("Exec", []string{"apply", "-f", applesFile, "-f", orangesFile, "-f", topicsFile}).Return("", nil)

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
		relPath := "../test_data/command/fn-with-existing-files"
		absPath, _ := filepath.Abs(relPath)
		applyCmd.SetArgs([]string{relPath})
		args := []string{"apply"}
		for _, f := range []string{"apples-function.yaml", "oranges-function.yaml", "echo-topics.yaml"} {
			args = append(append(args, "-f"), filepath.Join(absPath, f))
		}
		realKubeCtl.On("Exec", args).Return("", fmt.Errorf("Whoops"))

		err := applyCmd.Execute()
		Expect(err).To(MatchError("Whoops"))

	})

})
