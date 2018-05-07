package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/ghodss/yaml"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	projectriff_v1 "github.com/projectriff/riff/kubernetes-crds/pkg/apis/projectriff.io/v1alpha1"
	"github.com/projectriff/riff/riff-cli/pkg/options"
	"github.com/spf13/cobra"
)

var listInvokerArgs = []string{"get", "Invokers", "-o", "json"}

var _ = Describe("The init command", func() {
	var (
		oldCWD string

		commonRiffArgs []string
	)

	BeforeEach(func() {
		var err error

		oldCWD, err = os.Getwd()
		Expect(err).NotTo(HaveOccurred())

		commonRiffArgs = []string{"--force", "--useraccount", "rifftest"}
	})

	AfterEach(func() {
		os.Chdir(oldCWD)
	})

	Context("without an explict invoker", func() {

		It("should fail if no invokers are defined", func() {
			os.Chdir("../test_data/riff-init/no-invokers")

			invokers, err := stubInvokers("invokers/*.yaml")
			Expect(err).NotTo(HaveOccurred())
			rootCommand, _, _, _, err := setupInitTest(invokers)
			Expect(err).NotTo(HaveOccurred())

			rootCommand.SetArgs(append([]string{"init", "--artifact", "echo.sh"}, commonRiffArgs...))

			err = rootCommand.Execute()
			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(fmt.Errorf("Invokers must be installed, run `riff invokers apply --help` for help")))

			Expect(".").NotTo(HaveUnstagedChanges())
		})

		It("should fail if no invoker is specified", func() {
			os.Chdir("../test_data/riff-init/no-matching-invoker")

			invokers, err := stubInvokers("invokers/*.yaml")
			Expect(err).NotTo(HaveOccurred())
			rootCommand, _, _, _, err := setupInitTest(invokers)
			Expect(err).NotTo(HaveOccurred())

			rootCommand.SetArgs(append([]string{"init", "--artifact", "echo.sh"}, commonRiffArgs...))

			err = rootCommand.Execute()
			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(fmt.Errorf("The invoker must be specified. Pick one of: node")))

			Expect(".").NotTo(HaveUnstagedChanges())
		})

		It("ignores unknown flags", func() {
			os.Chdir("../test_data/riff-init/no-invokers")

			invokers, err := stubInvokers("invokers/*.yaml")
			Expect(err).NotTo(HaveOccurred())
			rootCommand, _, _, _, err := setupInitTest(invokers)
			Expect(err).NotTo(HaveOccurred())

			rootCommand.SetArgs(append([]string{"init", "--handler", "functions.FooFunc"}, commonRiffArgs...))

			err = rootCommand.Execute()
			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(fmt.Errorf("Invokers must be installed, run `riff invokers apply --help` for help")))

			Expect(".").NotTo(HaveUnstagedChanges())
		})

	})

	Context("with an explict invoker", func() {

		It("should fail if the invoker is not installed", func() {
			os.Chdir("../test_data/riff-init/no-matching-invoker")

			invokers, err := stubInvokers("invokers/*.yaml")
			Expect(err).NotTo(HaveOccurred())
			rootCommand, _, _, _, err := setupInitTest(invokers)
			Expect(err).NotTo(HaveOccurred())

			rootCommand.SetArgs(append([]string{"init", "command", "--artifact", "echo.sh"}, commonRiffArgs...))

			err = rootCommand.Execute()
			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(fmt.Errorf("The invoker must be specified. Pick one of: node")))

			Expect(".").NotTo(HaveUnstagedChanges())
		})

		It("ignores unknown flags", func() {
			os.Chdir("../test_data/riff-init/no-matching-invoker")

			invokers, err := stubInvokers("invokers/*.yaml")
			Expect(err).NotTo(HaveOccurred())
			rootCommand, _, _, _, err := setupInitTest(invokers)
			Expect(err).NotTo(HaveOccurred())

			rootCommand.SetArgs(append([]string{"init", "java", "--handler", "functions.FooFunc"}, commonRiffArgs...))

			err = rootCommand.Execute()
			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(fmt.Errorf("The invoker must be specified. Pick one of: node")))

			Expect(".").NotTo(HaveUnstagedChanges())
		})

		It("should detect the artifact for the specified invoker", func() {
			os.Chdir("../test_data/riff-init/matching-invoker")

			invokers, err := stubInvokers("invokers/*.yaml")
			Expect(err).NotTo(HaveOccurred())
			rootCommand, _, _, _, err := setupInitTest(invokers)
			Expect(err).NotTo(HaveOccurred())

			rootCommand.SetArgs(append([]string{"init", "node"}, commonRiffArgs...))

			err = rootCommand.Execute()
			Expect(err).NotTo(HaveOccurred())

			Expect(".").NotTo(HaveUnstagedChanges())
		})

		It("should require function name to be lower case", func() {
			os.Chdir("../test_data/riff-init/matching-invoker")

			invokers, err := stubInvokers("invokers/*.yaml")
			Expect(err).NotTo(HaveOccurred())
			rootCommand, _, _, _, err := setupInitTest(invokers)
			Expect(err).NotTo(HaveOccurred())

			name := "Echo"

			rootCommand.SetArgs(append([]string{"init", "node", "--name", name}, commonRiffArgs...))

			err = rootCommand.Execute()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(HavePrefix(fmt.Sprintf("function name %s is invalid", name)))

		})

		It("should ignore other matching invokers when explit invoker is selected", func() {
			os.Chdir("../test_data/riff-init/multiple-matching-invokers-with-one-selected")

			invokers, err := stubInvokers("invokers/*.yaml")
			Expect(err).NotTo(HaveOccurred())
			rootCommand, _, _, _, err := setupInitTest(invokers)
			Expect(err).NotTo(HaveOccurred())

			rootCommand.SetArgs(append([]string{"init", "python3", "--artifact", "echo.py"}, commonRiffArgs...))

			err = rootCommand.Execute()
			Expect(err).NotTo(HaveOccurred())

			Expect(".").NotTo(HaveUnstagedChanges())
		})

	})

	It("should allow for a custom function handler", func() {
		os.Chdir("../test_data/riff-init/matching-invoker-with-handler")

		invokers, err := stubInvokers("invokers/*.yaml")
		Expect(err).NotTo(HaveOccurred())
		rootCommand, _, _, _, err := setupInitTest(invokers)
		Expect(err).NotTo(HaveOccurred())

		rootCommand.SetArgs(append([]string{"init", "python3", "--artifact", "echo.py", "--handler", "customFuncName"}, commonRiffArgs...))

		err = rootCommand.Execute()
		Expect(err).NotTo(HaveOccurred())

		Expect(".").NotTo(HaveUnstagedChanges())
	})

	It("should allow for an output topic", func() {
		os.Chdir("../test_data/riff-init/matching-invoker-with-output-topic")

		invokers, err := stubInvokers("invokers/*.yaml")
		Expect(err).NotTo(HaveOccurred())
		rootCommand, _, _, _, err := setupInitTest(invokers)
		Expect(err).NotTo(HaveOccurred())

		rootCommand.SetArgs(append([]string{"init", "python3", "--artifact", "echo.py", "--output", "soundReflection"}, commonRiffArgs...))

		err = rootCommand.Execute()
		Expect(err).NotTo(HaveOccurred())

		Expect(".").NotTo(HaveUnstagedChanges())
	})

	It("should allow for a custom version", func() {
		os.Chdir("../test_data/riff-init/matching-invoker-with-version")

		invokers, err := stubInvokers("invokers/*.yaml")
		Expect(err).NotTo(HaveOccurred())
		rootCommand, _, _, _, err := setupInitTest(invokers)
		Expect(err).NotTo(HaveOccurred())

		rootCommand.SetArgs(append([]string{"init", "node", "--artifact", "echo.js", "--version", "latest"}, commonRiffArgs...))

		err = rootCommand.Execute()
		Expect(err).NotTo(HaveOccurred())

		Expect(".").NotTo(HaveUnstagedChanges())
	})

	It("should allow for a custom invoker version", func() {
		os.Chdir("../test_data/riff-init/matching-invoker-with-invoker-version")

		invokers, err := stubInvokers("invokers/*.yaml")
		Expect(err).NotTo(HaveOccurred())
		rootCommand, _, _, _, err := setupInitTest(invokers)
		Expect(err).NotTo(HaveOccurred())

		rootCommand.SetArgs(append([]string{"init", "node", "--artifact", "echo.js", "--invoker-version", "latest"}, commonRiffArgs...))

		err = rootCommand.Execute()
		Expect(err).NotTo(HaveOccurred())

		Expect(".").NotTo(HaveUnstagedChanges())
	})

})

func stubInvokers(invokerPattern string) ([]projectriff_v1.Invoker, error) {
	var invokers = []projectriff_v1.Invoker{}
	invokerPaths, err := filepath.Glob(invokerPattern)
	if err != nil {
		return nil, err
	}
	for _, invokerPath := range invokerPaths {
		bytes, err := ioutil.ReadFile(invokerPath)
		if err != nil {
			return nil, err
		}
		var invoker = projectriff_v1.Invoker{}
		err = yaml.Unmarshal(bytes, &invoker)
		if err != nil {
			return nil, err
		}
		invokers = append(invokers, invoker)
	}
	return invokers, nil
}

func setupInitTest(invokers []projectriff_v1.Invoker) (*cobra.Command, *cobra.Command, []*cobra.Command, *options.InitOptions, error) {
	rootCommand := Root()
	initCommand, initOptions := Init(invokers)
	initInvokerCommands, err := InitInvokers(invokers, initOptions)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	rootCommand.AddCommand(initCommand)
	initCommand.AddCommand(initInvokerCommands...)

	return rootCommand, initCommand, initInvokerCommands, initOptions, err
}
