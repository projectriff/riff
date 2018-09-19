package packs

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
)

const (
	CodeFailed      = 1
	CodeInvalidArgs = iota + 2
	CodeInvalidEnv
	CodeNotFound
	CodeFailedDetect
	CodeFailedBuild
	CodeFailedLaunch
	CodeFailedUpdate
)

type ErrorFail struct {
	Err    error
	Code   int
	Action []string
}

func (e *ErrorFail) Error() string {
	message := "failed to " + strings.Join(e.Action, " ")
	if e.Err == nil {
		return message
	}
	return fmt.Sprintf("%s: %s", message, e.Err)
}

func FailCode(code int, action ...string) error {
	return FailErrCode(nil, code, action...)
}

func FailErr(err error, action ...string) error {
	code := CodeFailed
	if err, ok := err.(*ErrorFail); ok {
		code = err.Code
	}
	return FailErrCode(err, code, action...)
}

func FailErrCode(err error, code int, action ...string) error {
	return &ErrorFail{Err: err, Code: code, Action: action}
}

func Exit(err error) {
	if err == nil {
		os.Exit(0)
	}
	log.Printf("Error: %s\n", err)
	if err, ok := err.(*ErrorFail); ok {
		os.Exit(err.Code)
	}
	os.Exit(CodeFailed)
}

func Run(name string, arg ...string) (string, error) {
	stdout, stderr := &bytes.Buffer{}, &bytes.Buffer{}
	cmd := exec.Command(name, arg...)
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	if err := cmd.Run(); err != nil {
		return "", FailErr(err, "run:", name, strings.Join(arg, " "), "\n", stderr.String())
	}
	return strings.TrimSpace(stdout.String()), nil
}

func BoolEnv(k string) bool {
	v := os.Getenv(k)
	return v == "true" || v == "1"
}
