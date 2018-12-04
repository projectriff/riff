package lifecycle

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/google/go-containerregistry/pkg/v1"
)

type Analyzer struct {
	Buildpacks []*Buildpack
	In         []byte
	Out, Err   io.Writer
}

func (a *Analyzer) Analyze(launchDir string, image v1.Image) error {
	config, err := a.getMetadata(image)
	if err != nil {
		return err
	}
	return a.AnalyzeConfig(launchDir, config)
}

func (a *Analyzer) AnalyzeConfig(launchDir string, config AppImageMetadata) error {
	buildpacks := a.buildpacks()
	for _, buildpack := range config.Buildpacks {
		if _, exist := buildpacks[buildpack.ID]; !exist {
			continue
		}
		for name, metadata := range buildpack.Layers {
			path := filepath.Join(launchDir, buildpack.ID, name+".toml")
			if err := writeTOML(path, metadata.Data); err != nil {
				return err
			}
		}
	}

	return nil
}

func (a *Analyzer) getMetadata(image v1.Image) (AppImageMetadata, error) {
	configFile, err := image.ConfigFile()
	if err != nil {
		return AppImageMetadata{}, err
	}
	jsonConfig := configFile.Config.Labels[MetadataLabel]
	if jsonConfig == "" {
		return AppImageMetadata{}, nil
	}

	config := AppImageMetadata{}
	if err := json.Unmarshal([]byte(jsonConfig), &config); err != nil {
		return AppImageMetadata{}, err
	}

	return config, err
}

func (a *Analyzer) buildpacks() map[string]struct{} {
	buildpacks := make(map[string]struct{}, len(a.Buildpacks))
	for _, b := range a.Buildpacks {
		buildpacks[b.ID] = struct{}{}
	}
	return buildpacks
}

func writeTOML(path string, data interface{}) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	fh, err := os.Create(path)
	if err != nil {
		return err
	}
	defer fh.Close()
	return toml.NewEncoder(fh).Encode(data)
}
