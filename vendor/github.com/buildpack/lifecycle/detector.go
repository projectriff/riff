package lifecycle

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os/exec"
	"path/filepath"
	"sync"
	"syscall"

	"github.com/BurntSushi/toml"
)

const (
	CodeDetectPass = iota
	CodeDetectError
	CodeDetectFail = 100
)

type Buildpack struct {
	ID       string `toml:"id"`
	Version  string `toml:"version"`
	Optional bool   `toml:"optional,omitempty"`
	Name     string `toml:"-"`
	Dir      string `toml:"-"`
}

func (bp *Buildpack) Detect(l *log.Logger, appDir string, in io.Reader, out io.Writer) int {
	path, err := filepath.Abs(filepath.Join(bp.Dir, "bin", "detect"))
	if err != nil {
		l.Print("Error: ", err)
		return CodeDetectError
	}
	stderr := &bytes.Buffer{}
	defer func() {
		if stderr.Len() > 0 {
			l.Print(stderr)
		}
	}()
	cmd := exec.Command(path)
	cmd.Dir = appDir
	cmd.Stdin = in
	cmd.Stdout = out
	cmd.Stderr = stderr
	if err := cmd.Run(); err != nil {
		if err, ok := err.(*exec.ExitError); ok {
			if status, ok := err.Sys().(syscall.WaitStatus); ok {
				return status.ExitStatus()
			}
		}
		l.Print("Error: ", err)
		return CodeDetectError
	}
	return CodeDetectPass
}

type BuildpackGroup struct {
	Buildpacks []*Buildpack `toml:"buildpacks"`
}

func (bg *BuildpackGroup) Detect(l *log.Logger, appDir string) (info []byte, group *BuildpackGroup, ok bool) {
	group = &BuildpackGroup{}
	detected := true
	summary := "Group:"
	info, codes := bg.pDetect(l, appDir)
	for i, code := range codes {
		if i > 0 {
			summary += " |"
		}
		switch code {
		case CodeDetectPass:
			summary += fmt.Sprintf(" %s: pass", bg.Buildpacks[i].Name)
			group.Buildpacks = append(group.Buildpacks, bg.Buildpacks[i])
		case CodeDetectFail:
			summary += fmt.Sprintf(" %s: fail", bg.Buildpacks[i].Name)
			detected = detected && bg.Buildpacks[i].Optional
		default:
			summary += fmt.Sprintf(" %s: error (%d)", bg.Buildpacks[i].Name, code)
			detected = detected && bg.Buildpacks[i].Optional
		}
	}
	detected = detected && len(group.Buildpacks) > 0
	l.Println(summary)
	return info, group, detected
}

func (bg *BuildpackGroup) pDetect(l *log.Logger, appDir string) (info []byte, codes []int) {
	codes = make([]int, len(bg.Buildpacks))
	wg := sync.WaitGroup{}
	defer wg.Wait()
	wg.Add(len(bg.Buildpacks))
	var lastIn io.ReadCloser
	for i := range bg.Buildpacks {
		in, out := io.Pipe()
		go func(i int, last io.ReadCloser) {
			defer wg.Done()
			defer out.Close()
			add := &bytes.Buffer{}
			if last != nil {
				defer last.Close()
				orig := &bytes.Buffer{}
				last := io.TeeReader(last, orig)
				codes[i] = bg.Buildpacks[i].Detect(l, appDir, last, add)
				ioutil.ReadAll(last)
				if codes[i] == CodeDetectPass {
					mergeTOML(l, out, orig, add)
				} else {
					mergeTOML(l, out, orig)
				}
			} else {
				codes[i] = bg.Buildpacks[i].Detect(l, appDir, nil, add)
				if codes[i] == CodeDetectPass {
					mergeTOML(l, out, add)
				}
			}
		}(i, lastIn)
		lastIn = in
	}
	if lastIn != nil {
		defer lastIn.Close()
		if i, err := ioutil.ReadAll(lastIn); err != nil {
			l.Print("Warning: ", err)
		} else {
			info = i
		}
	}
	return info, codes
}

func mergeTOML(l *log.Logger, out io.Writer, in ...io.Reader) {
	result := map[string]interface{}{}
	for _, r := range in {
		var m map[string]interface{}
		if _, err := toml.DecodeReader(r, &m); err != nil {
			l.Print("Warning: ", err)
			continue
		}
		for k, v := range m {
			result[k] = v
		}
	}
	if err := toml.NewEncoder(out).Encode(result); err != nil {
		l.Print("Warning: ", err)
	}
}

type BuildpackOrder []BuildpackGroup

func (bo BuildpackOrder) Detect(l *log.Logger, appDir string) ([]byte, *BuildpackGroup) {
	for i := range bo {
		if info, group, ok := bo[i].Detect(l, appDir); ok {
			return info, group
		}
	}
	return nil, nil
}
