package cmd

import (
	"fmt"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/projectriff/riff/riff-cli/pkg/kubectl"
	"github.com/spf13/cobra"
)

var _ = Describe("The delete command", func() {

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
		deleteCmd     *cobra.Command
		args          []string
	)

	BeforeEach(func() {
		var err error

		oldCWD, err = os.Getwd()
		Expect(err).NotTo(HaveOccurred())

		realKubeCtl = new(kubectl.MockKubeCtl)
		dryRunKubeCtl = new(kubectl.MockKubeCtl)

		deleteCmd, _ = Delete(realKubeCtl, dryRunKubeCtl)
		args = []string{}

	})

	AfterEach(func() {
		os.Chdir(oldCWD)
	})

	Context("with no --name provided", func() {
		BeforeEach(func() {
			os.Chdir("../test_data/command/echo")
		})

		It("should delete the function based on dirname", func() {

			realKubeCtl.On("Exec", []string{"delete", "functions.projectriff.io", "echo"}).Return("", nil)

			err := deleteCmd.Execute()
			Expect(err).NotTo(HaveOccurred())

		})

		It("should delete the function and topic when run with --all", func() {

			deleteCmd.SetArgs([]string{"--all"})
			realKubeCtl.On("Exec", []string{"delete", "functions.projectriff.io", "echo"}).Return("", nil)

			realKubeCtl.On("Exec", []string{"get", "functions.projectriff.io", "echo", "-o", "json"}).Return(canned_kubectl_get_response, nil)
			realKubeCtl.On("Exec", []string{"delete", "topics.projectriff.io", "myInputTopic"}).Return("", nil)
			realKubeCtl.On("Exec", []string{"delete", "topics.projectriff.io", "myOutputTopic"}).Return("", nil)

			err := deleteCmd.Execute()
			Expect(err).NotTo(HaveOccurred())

		})

		Context("when --namespace is set", func() {
			BeforeEach(func() {
				args = append(args, "--namespace", "my-ns")
			})
			It("should delete the function based on dirname", func() {

				deleteCmd.SetArgs(args)
				realKubeCtl.On("Exec", []string{"delete", "functions.projectriff.io", "echo", "--namespace", "my-ns"}).Return("", nil)

				err := deleteCmd.Execute()
				Expect(err).NotTo(HaveOccurred())

			})

			It("should delete the function and topic when run with --all", func() {

				args = append(args, "--all")
				deleteCmd.SetArgs(args)
				realKubeCtl.On("Exec", []string{"delete", "functions.projectriff.io", "echo", "--namespace", "my-ns"}).Return("", nil)

				realKubeCtl.On("Exec", []string{"get", "--namespace", "my-ns", "functions.projectriff.io", "echo", "-o", "json"}).Return(canned_kubectl_get_response, nil)
				realKubeCtl.On("Exec", []string{"delete", "topics.projectriff.io", "myInputTopic", "--namespace", "my-ns"}).Return("", nil)
				realKubeCtl.On("Exec", []string{"delete", "topics.projectriff.io", "myOutputTopic", "--namespace", "my-ns"}).Return("", nil)

				err := deleteCmd.Execute()
				Expect(err).NotTo(HaveOccurred())

			})

		})
	})

	Context("when --name is provided", func() {
		BeforeEach(func() {
			args = append(args, "--name", "my-function")
		})

		It("should delete the function based on name", func() {
			deleteCmd.SetArgs(args)

			realKubeCtl.On("Exec", []string{"delete", "functions.projectriff.io", "my-function"}).Return("", nil)

			err := deleteCmd.Execute()
			Expect(err).NotTo(HaveOccurred())

		})

		It("should delete the function and topic when run with --all", func() {

			args = append(args, "--all")
			deleteCmd.SetArgs(args)

			realKubeCtl.On("Exec", []string{"delete", "functions.projectriff.io", "my-function"}).Return("", nil)

			realKubeCtl.On("Exec", []string{"get", "functions.projectriff.io", "my-function", "-o", "json"}).Return(canned_kubectl_get_response, nil)
			realKubeCtl.On("Exec", []string{"delete", "topics.projectriff.io", "myInputTopic"}).Return("", nil)
			realKubeCtl.On("Exec", []string{"delete", "topics.projectriff.io", "myOutputTopic"}).Return("", nil)

			err := deleteCmd.Execute()
			Expect(err).NotTo(HaveOccurred())

		})

		Context("when --namespace is set", func() {
			BeforeEach(func() {
				args = append(args, "--namespace", "my-ns")
			})
			It("should delete the function based on name", func() {

				deleteCmd.SetArgs(args)
				realKubeCtl.On("Exec", []string{"delete", "functions.projectriff.io", "my-function", "--namespace", "my-ns"}).Return("", nil)

				err := deleteCmd.Execute()
				Expect(err).NotTo(HaveOccurred())

			})

			It("should delete the function and topic when run with --all", func() {

				args = append(args, "--all")
				deleteCmd.SetArgs(args)
				realKubeCtl.On("Exec", []string{"delete", "functions.projectriff.io", "my-function", "--namespace", "my-ns"}).Return("", nil)

				realKubeCtl.On("Exec", []string{"get", "--namespace", "my-ns", "functions.projectriff.io", "my-function", "-o", "json"}).Return(canned_kubectl_get_response, nil)
				realKubeCtl.On("Exec", []string{"delete", "topics.projectriff.io", "myInputTopic", "--namespace", "my-ns"}).Return("", nil)
				realKubeCtl.On("Exec", []string{"delete", "topics.projectriff.io", "myOutputTopic", "--namespace", "my-ns"}).Return("", nil)

				err := deleteCmd.Execute()
				Expect(err).NotTo(HaveOccurred())

			})

		})
	})

	It("should report kubectl errors", func() {
		deleteCmd.SetArgs([]string{"--name", "whatever"})

		realKubeCtl.On("Exec", []string{"delete", "functions.projectriff.io", "whatever"}).Return("", fmt.Errorf("Whoops"))

		err := deleteCmd.Execute()
		Expect(err).To(MatchError("Whoops"))

	})

	It("should not use the real kubectl client when using --dry-run", func() {
		deleteCmd.SetArgs([]string{"--name", "whatever", "--dry-run"})

		dryRunKubeCtl.On("Exec", []string{"delete", "functions.projectriff.io", "whatever"}).Return("", nil)

		err := deleteCmd.Execute()
		Expect(err).NotTo(HaveOccurred())

	})

})
