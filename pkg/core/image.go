/*
 * Copyright 2018 The original author or authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package core

import (
	"bytes"
	"fmt"
	"github.com/ghodss/yaml"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
)

const outputFilePermissions = 0644

type RelocateImageOptions struct {
	YAML     string
	Manifest string
	Output   string

	Registry     string
	RegistryUser string
	Images       string
}

func (c *client) RelocateImage(options RelocateImageOptions) error {
	imageMapper, err := createImageMapper(options)
	if err != nil {
		return err
	}

	if options.YAML != "" {
		_, err = relocateYAMLFile(options.YAML, imageMapper, "", options.Output)
		return err
	}
	return relocateManifest(options.Manifest, imageMapper, options.Output)
}

func createImageMapper(options RelocateImageOptions) (*imageMapper, error) {
	imageManifest, err := NewImageManifest(options.Images)
	if err != nil {
		return nil, err
	}

	imageMapper, err := newImageMapper(options.Registry, options.RegistryUser, imageManifest.Images)
	if err != nil {
		return nil, err
	}

	return imageMapper, nil
}

func relocateYAMLFile(yamlFile string, mapper *imageMapper, cwd string, outputPath string) (string, error) {
	y, err := readYAMLFile(yamlFile, cwd)
	if err != nil {
		return "", err
	}

	output := outputFile(outputPath, yamlFile)
	return filepath.Base(output), ioutil.WriteFile(output, mapper.mapImages(y), outputFilePermissions)
}

func readYAMLFile(yamlFile string, cwd string) ([]byte, error) {
	u, err := url.Parse(yamlFile)
	if err != nil {
		return nil, err
	}
	if u.IsAbs() {
		if u.Scheme == "file" {
			return ioutil.ReadFile(u.Path)
		}
		// FIXME: Error: Get https:///storage.googleapis.com/knative-releases/serving/previous/v20180828-7c20145/istio.yaml: http: no Host in request URL
		return downloadFile(u.String())
	}

	if !filepath.IsAbs(yamlFile) {
		yamlFile = filepath.Join(cwd, yamlFile)
	}
	return ioutil.ReadFile(yamlFile)
}

func downloadFile(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(resp.Body)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func relocateManifest(manifestPath string, mapper *imageMapper, outputPath string) error {
	if !isDirectory(outputPath) {
		return fmt.Errorf("Invalid or non-existent output directory '%s'", outputPath)
	}

	manifest, err := NewManifest(manifestPath)
	if err != nil {
		return err
	}

	manifestDir := filepath.Dir(manifestPath)

	outputManifest := &Manifest{
		ManifestVersion: manifest.ManifestVersion,
	}

	outputManifest.Istio, err = relocateYamls(manifest.Istio, manifestDir, mapper, outputPath)
	if err != nil {
		return err
	}
	outputManifest.Knative, err = relocateYamls(manifest.Knative, manifestDir, mapper, outputPath)
	if err != nil {
		return err
	}
	outputManifest.Namespace, err = relocateYamls(manifest.Namespace, manifestDir, mapper, outputPath)
	if err != nil {
		return err
	}

	outputManifestPath := filepath.Join(outputPath, "manifest.yaml")
	outputManifestBytes, err := yaml.Marshal(&outputManifest)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(outputManifestPath, outputManifestBytes, outputFilePermissions)
}

func relocateYamls(yamls []string, manifestDir string, mapper *imageMapper, outputPath string) ([]string, error) {
	rel := []string{}
	for _, y := range yamls {
		// FIXME so this works for both relative paths and URLs
		//err := relocateYAMLFile(filepath.Join(manifestDir, y), mapper, outputPath)
		yRel, err := relocateYAMLFile(y, mapper, manifestDir, outputPath)
		if err != nil {
			return []string{}, err
		}
		rel = append(rel, yRel)
	}
	return rel, nil
}

func outputFile(outputPath string, yamlFile string) string {
	if isDirectory(outputPath) {
		return filepath.Join(outputPath, filepath.Base(yamlFile))
	}
	return outputPath
}

func isDirectory(path string) bool {
	f, err := os.Stat(path)
	return err == nil && f.IsDir()
}
