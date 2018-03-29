package cmd

import (
	"context"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/onsi/gomega/types"
	"github.com/projectriff/riff/riff-cli/pkg/ioutils"

	"fmt"
)

func HaveUnstagedChanges() types.GomegaMatcher {
	return &unstagedChangesMatcher{}
}

type unstagedChangesMatcher struct{}

func (matcher *unstagedChangesMatcher) Match(actual interface{}) (success bool, err error) {
	// Create a new context and add a timeout to it
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel() // The cancel should be deferred so resources are cleaned up

	// Create the command with our context
	cmd := exec.CommandContext(ctx, "git", "diff", "--exit-code", actual.(string))
	err = cmd.Run()

	// We want to check the context error to see if the timeout was executed.
	// The error returned by cmd.Output() will be OS specific based on what
	// happens when a process is killed.
	if ctx.Err() == context.DeadlineExceeded {
		ioutils.Error("Command timed out")
		return false, ctx.Err()
	}

	if err != nil {
		return true, nil
	}
	return false, nil
}

func (matcher *unstagedChangesMatcher) FailureMessage(actual interface{}) (message string) {
	path, _ := filepath.Abs(actual.(string))
	return fmt.Sprintf("Expected\n\t%#v\nto have unstaged git changes", path)
}

func (matcher *unstagedChangesMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	path, _ := filepath.Abs(actual.(string))
	return fmt.Sprintf("Expected\n\t%#v\nnot to have unstaged git changes", path)
}
