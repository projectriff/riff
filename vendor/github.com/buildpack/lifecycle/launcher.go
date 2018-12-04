package lifecycle

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
)

type Launcher struct {
	DefaultProcessType string
	DefaultLaunchDir   string
	Processes          []Process
	Buildpacks         []string
	Exec               func(argv0 string, argv []string, envv []string) error
}

func (l *Launcher) Launch(executable, startCommand string) error {
	env := &Env{
		Getenv:  os.Getenv,
		Setenv:  os.Setenv,
		Environ: os.Environ,
		Map:     POSIXLaunchEnv,
	}
	if err := l.eachDir(l.DefaultLaunchDir, func(bp string) error {
		if bp == "app" {
			return nil
		}
		bpPath := filepath.Join(l.DefaultLaunchDir, bp)
		return l.eachDir(bpPath, func(layer string) error {
			return env.AddRootDir(filepath.Join(bpPath, layer))
		})
	}); err != nil {
		return errors.Wrap(err, "modify env")
	}
	if err := os.Chdir(filepath.Join(l.DefaultLaunchDir, "app")); err != nil {
		return errors.Wrap(err, "change to app directory")
	}

	startCommand, err := l.processFor(startCommand)
	if err != nil {
		return errors.Wrap(err, "determine start command")
	}

	launcher, err := l.profileD()
	if err != nil {
		return errors.Wrap(err, "determine profile")
	}

	if err := l.Exec("/bin/bash", []string{
		"bash", "-c",
		launcher, executable,
		startCommand,
	}, os.Environ()); err != nil {
		return errors.Wrap(err, "exec")
	}
	return nil
}

func (l *Launcher) profileD() (string, error) {
	var out []string

	appendIfFile := func(path string) error {
		fi, err := os.Stat(path)
		if os.IsNotExist(err) {
			return nil
		}
		if err != nil {
			return err
		}
		if !fi.IsDir() {
			out = append(out, fmt.Sprintf(`source "%s"`, path))
		}
		return nil
	}

	for _, bp := range l.Buildpacks {
		scripts, err := filepath.Glob(filepath.Join(l.DefaultLaunchDir, bp, "*", "profile.d", "*"))
		if err != nil {
			return "", err
		}
		for _, script := range scripts {
			if err := appendIfFile(script); err != nil {
				return "", err
			}
		}
	}

	if err := appendIfFile(filepath.Join(l.DefaultLaunchDir, "app", ".profile")); err != nil {
		return "", err
	}

	out = append(out, `exec bash -c "$@"`)
	return strings.Join(out, "\n"), nil
}

func (l *Launcher) processFor(cmd string) (string, error) {
	if cmd == "" {
		if process, ok := l.findProcessType(l.DefaultProcessType); ok {
			return process, nil
		}

		return "", fmt.Errorf("process type %s was not found", l.DefaultProcessType)
	}

	if process, ok := l.findProcessType(cmd); ok {
		return process, nil
	}

	return cmd, nil
}

func (l *Launcher) findProcessType(kind string) (string, bool) {
	for _, p := range l.Processes {
		if p.Type == kind {
			return p.Command, true
		}
	}

	return "", false
}

func (*Launcher) eachDir(dir string, fn func(file string) error) error {
	files, err := ioutil.ReadDir(dir)
	if os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return err
	}
	for _, f := range files {
		if !f.IsDir() {
			continue
		}
		if err := fn(f.Name()); err != nil {
			return err
		}
	}
	return nil
}
