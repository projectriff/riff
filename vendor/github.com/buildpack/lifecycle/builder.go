package lifecycle

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"sort"

	"github.com/BurntSushi/toml"
)

type Builder struct {
	PlatformDir string
	Buildpacks  []*Buildpack
	In          []byte
	Out, Err    io.Writer
}

type BuildEnv interface {
	AddRootDir(baseDir string) error
	AddEnvDir(envDir string) error
	List() []string
}

type Process struct {
	Type    string `toml:"type"`
	Command string `toml:"command"`
}

type LaunchTOML struct {
	Processes []Process `toml:"processes"`
}

type BuildMetadata struct {
	Processes  []Process `toml:"processes"`
	Buildpacks []string  `toml:"buildpacks"`
}

func (b *Builder) Build(cacheDir, launchDir string, env BuildEnv) (*BuildMetadata, error) {
	procMap := processMap{}
	var buildpackIDs []string
	for _, bp := range b.Buildpacks {
		bpLaunchDir := filepath.Join(launchDir, bp.ID)
		bpCacheDir := filepath.Join(cacheDir, bp.ID)
		buildpackIDs = append(buildpackIDs, bp.ID)
		if err := os.MkdirAll(bpLaunchDir, 0777); err != nil {
			return nil, err
		}
		if err := os.MkdirAll(bpCacheDir, 0777); err != nil {
			return nil, err
		}
		buildPath, err := filepath.Abs(filepath.Join(bp.Dir, "bin", "build"))
		if err != nil {
			return nil, err
		}
		cmd := exec.Command(buildPath, b.PlatformDir, bpCacheDir, bpLaunchDir)
		cmd.Env = env.List()
		cmd.Dir = filepath.Join(launchDir, "app")
		cmd.Stdin = bytes.NewBuffer(b.In)
		cmd.Stdout = b.Out
		cmd.Stderr = b.Err
		if err := cmd.Run(); err != nil {
			return nil, err
		}
		if err := setupEnv(env, bpCacheDir); err != nil {
			return nil, err
		}
		var launch LaunchTOML
		tomlPath := filepath.Join(bpLaunchDir, "launch.toml")
		if _, err := toml.DecodeFile(tomlPath, &launch); os.IsNotExist(err) {
			continue
		} else if err != nil {
			return nil, err
		}
		procMap.add(launch.Processes)
	}

	return &BuildMetadata{
		Processes:  procMap.list(),
		Buildpacks: buildpackIDs,
	}, nil
}

type DevelopTOML struct {
	Processes []Process `toml:"processes"`
}

type DevelopMetadata DevelopTOML

func (b *Builder) Develop(appDir, cacheDir string, env BuildEnv) (*DevelopMetadata, error) {
	procMap := processMap{}
	for _, bp := range b.Buildpacks {
		bpCacheDir := filepath.Join(cacheDir, bp.ID)
		if err := os.MkdirAll(bpCacheDir, 0777); err != nil {
			return nil, err
		}
		developPath, err := filepath.Abs(filepath.Join(bp.Dir, "bin", "develop"))
		if err != nil {
			return nil, err
		}
		cmd := exec.Command(developPath, b.PlatformDir, bpCacheDir)
		cmd.Env = env.List()
		cmd.Dir = appDir
		cmd.Stdout = b.Out
		cmd.Stderr = b.Err
		if err := cmd.Run(); err != nil {
			return nil, err
		}
		if err := setupEnv(env, bpCacheDir); err != nil {
			return nil, err
		}
		var develop DevelopTOML
		tomlPath := filepath.Join(bpCacheDir, "develop.toml")
		if _, err := toml.DecodeFile(tomlPath, &develop); os.IsNotExist(err) {
			continue
		} else if err != nil {
			return nil, err
		}
		procMap.add(develop.Processes)
	}
	return &DevelopMetadata{
		Processes: procMap.list(),
	}, nil
}

func setupEnv(env BuildEnv, cacheDir string) error {
	cacheFiles, err := ioutil.ReadDir(cacheDir)
	if err != nil {
		return err
	}
	if err := eachDir(cacheFiles, func(layer os.FileInfo) error {
		return env.AddRootDir(filepath.Join(cacheDir, layer.Name()))
	}); err != nil {
		return err
	}
	return eachDir(cacheFiles, func(layer os.FileInfo) error {
		return env.AddEnvDir(filepath.Join(cacheDir, layer.Name(), "env"))
	})
}

func eachDir(files []os.FileInfo, fn func(os.FileInfo) error) error {
	for _, f := range files {
		if !f.IsDir() {
			continue
		}
		if err := fn(f); err != nil {
			return err
		}
	}
	return nil
}

type processMap map[string]Process

func (m processMap) add(l []Process) {
	for _, proc := range l {
		m[proc.Type] = proc
	}
}

func (m processMap) list() []Process {
	var keys []string
	for key := range m {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	procs := []Process{}
	for _, key := range keys {
		procs = append(procs, m[key])
	}
	return procs
}
