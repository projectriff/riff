package img

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/google/go-containerregistry/pkg/v1/tarball"
)

func Append(base v1.Image, tar string) (v1.Image, v1.Layer, error) {
	layer, err := tarball.LayerFromFile(tar)
	if err != nil {
		return nil, nil, err
	}
	image, err := mutate.AppendLayers(base, layer)
	if err != nil {
		return nil, nil, err
	}
	return image, layer, nil
}

type ImageFinder func(labels map[string]string) (v1.Image, error)

func Rebase(orig v1.Image, newBase v1.Image, oldBaseFinder ImageFinder) (v1.Image, error) {
	origConfig, err := orig.ConfigFile()
	if err != nil {
		return nil, err
	}
	oldBase, err := oldBaseFinder(origConfig.Config.Labels)
	if err != nil {
		return nil, err
	}
	image, err := mutate.Rebase(orig, oldBase, newBase, nil)
	if err != nil {
		return nil, err
	}
	return image, nil
}

func TopLayerDiffID(image v1.Image) (v1.Hash, error) {
	layers, err := image.Layers()
	if err != nil {
		return v1.Hash{}, err
	}
	return layers[len(layers)-1].DiffID()
}

func Label(image v1.Image, k, v string) (v1.Image, error) {
	configFile, err := image.ConfigFile()
	if err != nil {
		return nil, err
	}
	config := *configFile.Config.DeepCopy()
	if config.Labels == nil {
		config.Labels = map[string]string{}
	}
	config.Labels[k] = v
	return mutate.Config(image, config)
}

func Env(image v1.Image, k, v string) (v1.Image, error) {
	configFile, err := image.ConfigFile()
	if err != nil {
		return nil, err
	}
	config := *configFile.Config.DeepCopy()
	for i, e := range config.Env {
		parts := strings.Split(e, "=")
		if parts[0] == k {
			config.Env[i] = fmt.Sprintf("%s=%s", k, v)
			return mutate.Config(image, config)
		}
	}
	config.Env = append(config.Env, fmt.Sprintf("%s=%s", k, v))
	return mutate.Config(image, config)
}

func SetupCredHelpers(refs ...string) error {
	dockerPath := filepath.Join(os.Getenv("HOME"), ".docker")
	configPath := filepath.Join(dockerPath, "config.json")
	config := map[string]interface{}{}
	credHelpers := map[string]string{}
	config["credHelpers"] = credHelpers
	if f, err := os.Open(configPath); err == nil {
		err := json.NewDecoder(f).Decode(&config)
		if f.Close(); err != nil {
			return err
		}
	} else if !os.IsNotExist(err) {
		return err
	}
	added := false
	for _, refStr := range refs {
		ref, err := name.ParseReference(refStr, name.WeakValidation)
		if err != nil {
			return err
		}
		registry := ref.Context().RegistryStr()
		for _, ch := range []struct {
			domain string
			helper string
		}{
			{"([.]|^)gcr[.]io$", "gcr"},
			{"[.]amazonaws[.]", "ecr-login"},
			{"([.]|^)azurecr[.]io$", "acr"},
		} {
			match, err := regexp.MatchString("(?i)"+ch.domain, registry)
			if err != nil || !match {
				continue
			}
			credHelpers[registry] = ch.helper
			added = true
		}
	}
	if !added {
		return nil
	}
	if err := os.MkdirAll(dockerPath, 0777); err != nil {
		return err
	}
	f, err := os.Create(configPath)
	if err != nil {
		return err
	}
	defer f.Close()
	return json.NewEncoder(f).Encode(config)
}
