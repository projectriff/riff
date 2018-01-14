package kubectl

import (
	"os/exec"
	"time"
	"context"
	"github.com/projectriff/riff-cli/ioutils"
)

func QueryForString(cmdArgs []string) (string, error) {
	out, err := QueryForBytes(cmdArgs)
	return string(out), err
}

func QueryForBytes(cmdArgs []string) ([]byte, error) {

	// Create a new context and add a timeout to it
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel() // The cancel should be deferred so resources are cleaned up

	cmdName := "kubectl"
	// Create the command with our context
	cmd := exec.CommandContext(ctx, cmdName, cmdArgs...)

	// This time we can simply use Output() to get the result.
	out, err := cmd.Output()

	// We want to check the context error to see if the timeout was executed.
	// The error returned by cmd.Output() will be OS specific based on what
	// happens when a process is killed.
	if ctx.Err() == context.DeadlineExceeded {
		ioutils.Error("Command timed out")
		return nil, ctx.Err()
	}

	return out, err
}
