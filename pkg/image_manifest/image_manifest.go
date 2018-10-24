package image_manifest

import (
	"fmt"
	"io/ioutil"

	"github.com/ghodss/yaml"
	"github.com/projectriff/riff/pkg/image"
)

const (
	imageManifestVersion_0_1 = "0.1"
	outputFilePermissions    = 0644
)

// ImageManifest defines the image names found in YAML files of system components.
type ImageManifest struct {
	ManifestVersion string
	Images          map[image.Name]image.Id
}

// jsonImageManifest is used for (un)marshalling ImageManifest since JSON maps must have string keys and consequently
// Go maps that can be (un)marshalled must have string keys.
type jsonImageManifest struct {
	ManifestVersion string            `json:"manifestVersion"`
	Images          map[string]string `json:images`
}

func NewImageManifest() *ImageManifest {
	return &ImageManifest{
		ManifestVersion: imageManifestVersion_0_1,
		Images:          make(map[image.Name]image.Id),
	}
}

func PrimeImageManifest(images []string) (*ImageManifest, error) {
	im := NewImageManifest()
	for _, i := range images {
		ref, err := image.NewName(i)
		if err != nil {
			return nil, err
		}
		im.Images[ref] = image.EmptyId
	}
	return im, nil
}

func LoadImageManifest(path string) (*ImageManifest, error) {
	var jm jsonImageManifest
	yamlFile, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("error reading image manifest file: %v", err)
	}

	err = yaml.Unmarshal(yamlFile, &jm)
	if err != nil {
		return nil, fmt.Errorf("error parsing image manifest file: %v", err)
	}

	if jm.ManifestVersion != imageManifestVersion_0_1 {
		return nil, fmt.Errorf("image manifest has unsupported version: %s", jm.ManifestVersion)
	}

	if jm.Images == nil {
		return nil, fmt.Errorf("image manifest is incomplete: images map is missing: %#v", jm)
	}

	m := NewImageManifest()
	m.ManifestVersion = jm.ManifestVersion

	err = m.addImages(jm.Images)
	if err != nil {
		return nil, err
	}

	return m, nil
}

type Filter func(name image.Name, dig image.Id) (image.Name, image.Id, error)

func (m *ImageManifest) FilterCopy(filter Filter) (*ImageManifest, error) {
	newManifest := NewImageManifest()

	for name, dig := range m.Images {
		newName, newDig, err := filter(name, dig)
		if err != nil {
			return nil, err
		}
		newManifest.Images[newName] = newDig
	}
	return newManifest, nil
}

func (m *ImageManifest) Save(path string) error {
	jm := jsonImageManifest{
		ManifestVersion: m.ManifestVersion,
		Images:          make(map[string]string),
	}
	for k, v := range m.Images {
		jm.Images[k.String()] = v.String()
	}

	bytes, err := yaml.Marshal(&jm)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(path, bytes, outputFilePermissions)
}

func (m *ImageManifest) addImages(images map[string]string) error {
	for i, d := range images {
		err := m.addImage(i, d)
		if err != nil {
			return err
		}
	}

	return nil
}

func (m *ImageManifest) addImage(i string, d string) error {
	in, err := image.NewName(i)
	if err != nil {
		return err
	}
	m.Images[in] = image.NewId(d)
	return nil
}
