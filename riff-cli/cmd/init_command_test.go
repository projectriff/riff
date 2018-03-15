package cmd

import (
	"github.com/projectriff/riff/riff-cli/pkg/options"
	"github.com/spf13/cobra"
	"github.com/projectriff/riff/riff-cli/pkg/osutils"
	"os"
	"testing"
	"github.com/stretchr/testify/assert"
	"fmt"
)

func TestInitCommandExplicitPathAndLang(t *testing.T) {
	rootCmd, initOptions, _ := setupInitTest()
	as := assert.New(t)
	rootCmd.SetArgs([]string{"init", "command", "--dry-run", "-f", "../test_data/command/echo", "-v", "0.0.1-snapshot"})

	_, err := rootCmd.ExecuteC()
	as.NoError(err)
	as.Equal("../test_data/command/echo", initOptions.FilePath)
}

func TestInitCommandExplicitPath(t *testing.T) {
	rootCmd, initOptions, _ := setupInitTest()
	as := assert.New(t)
	rootCmd.SetArgs([]string{"init", "--dry-run", "-f", "../test_data/command/echo", "-v", "0.0.1-snapshot"})

	_, err := rootCmd.ExecuteC()
	as.NoError(err)
	as.Equal("../test_data/command/echo", initOptions.FilePath)
}

func TestInitCommandImplicitPath(t *testing.T) {
	rootCmd, initOptions, _ := setupInitTest()
	as := assert.New(t)
	rootCmd.SetArgs([]string{"init", "--dry-run", "../test_data/command/echo", "-v", "0.0.1-snapshot"})

	_, err := rootCmd.ExecuteC()
	as.NoError(err)
	as.Equal("../test_data/command/echo", initOptions.FilePath)
}

func TestInitJavaWithVersion(t *testing.T) {

	rootCmd, _, _ := setupInitTest()

	currentdir := osutils.GetCWD()
	path := osutils.Path("../test_data/java")
	os.Chdir(path)
	as := assert.New(t)
	rootCmd.SetArgs([]string{"init", "java", "--dry-run", "-a", "target/upper-1.0.0.jar", "--handler", "function.Upper"})
	_, err := rootCmd.ExecuteC()
	as.NoError(err)
	os.Chdir(currentdir)
}

func TestInitJava(t *testing.T) {
	rootCmd, _, _ := setupInitTest()
	currentdir := osutils.GetCWD()
	path := osutils.Path("../test_data/java")
	os.Chdir(path)
	as := assert.New(t)
	rootCmd.SetArgs([]string{"init", "java", "--dry-run", "-a", "target/dummy.jar", "--handler", "function.Dummy"})

	_, err := rootCmd.ExecuteC()
	fmt.Printf("%v\n", err)
	as.NoError(err)
	os.Chdir(currentdir)
}

func setupInitTest() (*cobra.Command, *options.InitOptions, map[string]*cobra.Command) {
	rootCmd := Root()

	initCmd, initOptions := Init()
	initJavaCmd, _ := InitJava(initOptions)
	initNodeCmd, _ := InitNode(initOptions)
	initPythonCmd, _ := InitPython(initOptions)
	initShellCmd, _ := InitCommand(initOptions)
	initGoCmd, _ := InitGo(initOptions)

	initCmd.AddCommand(
		initJavaCmd,
		initGoCmd,
		initShellCmd,
		initPythonCmd,
		initNodeCmd,
	)

	rootCmd.AddCommand(
		initCmd,
	)

	commands := make(map[string]*cobra.Command)
	registerCommand(commands, rootCmd)
	registerCommand(commands, initCmd)
	registerCommand(commands, initJavaCmd)
	registerCommand(commands, initNodeCmd)
	registerCommand(commands, initPythonCmd)
	registerCommand(commands, initShellCmd)
	registerCommand(commands, initGoCmd)

	return rootCmd, initOptions, commands
}

func registerCommand(commands map[string]*cobra.Command, command *cobra.Command) {
	commands[command.Name()] = command
}
