package logutil

import (
	"fmt"
	"io"
	"log"
	"os"
)

var exitFn = os.Exit

type logger struct {
	parent *log.Logger
	out    io.Writer
	cmp    string
}

func Default() Log {
	return New(log.New(os.Stderr, "", log.LstdFlags), os.Stderr)
}

func New(parent *log.Logger, out io.Writer) Log {
	return &logger{parent, out, ""}
}

func (l *logger) WithComponent(cmp string) Log {
	parent := log.New(l.out, l.parent.Prefix(), l.parent.Flags())
	return &logger{parent, l.out, cmp}
}

func (l *logger) Trace(format string, args ...interface{}) string {
	name := fmt.Sprintf(format, args...)
	l.output("ENTER", name)
	return name
}

func (l *logger) Un(name string) {
	l.output("LEAVE", name)
}

func (l *logger) Debugf(format string, args ...interface{}) {
	l.output("DEBUG", format, args...)
}

func (l *logger) Infof(format string, args ...interface{}) {
	l.output("INFO", format, args...)
}

func (l *logger) Warnf(format string, args ...interface{}) {
	l.output("WARN", format, args...)
}

func (l *logger) Errorf(format string, args ...interface{}) {
	l.output("ERROR", format, args...)
}

func (l *logger) Fatalf(format string, args ...interface{}) {
	l.output("FATAL", format, args...)
	exitFn(1)
}

func (l *logger) ErrWarn(err error, format string, args ...interface{}) error {
	l.output("WARN", "[ %v ] "+format, append([]interface{}{err}, args...)...)
	return err
}

func (l *logger) ErrFatal(err error, format string, args ...interface{}) error {
	l.output("FATAL", "[ %v ] "+format, append([]interface{}{err}, args...)...)
	exitFn(1)
	return err
}

func (l *logger) Err(err error, format string, args ...interface{}) error {
	l.output("ERROR", "[ %v ] "+format, append([]interface{}{err}, args...)...)
	return err
}

func (l *logger) output(level string, format string, args ...interface{}) {
	cmp := l.cmp
	if cmp != "" {
		cmp = "<" + cmp + "> "
	}
	message := fmt.Sprintf("%-6s %v"+format, append([]interface{}{level + ":", cmp}, args...)...)
	l.parent.Output(3, message)
}

func setExitFn(fn func(int)) {
	exitFn = fn
}
