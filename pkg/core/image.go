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
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/projectriff/riff/pkg/fileutils"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"

	"github.com/ghodss/yaml"
)

const (
	outputFilePermissions = 0644
	outputDirPermissions  = 0755
)

type RelocateImagesOptions struct {
	SingleFile string
	Manifest   string
	Output     string

	Registry     string
	RegistryUser string
	Images       string
	Flatten      bool
}

func (c *client) RelocateImages(options RelocateImagesOptions) error {
	imageMapper, err := createImageMapper(options)
	if err != nil {
		return err
	}

	if options.SingleFile != "" {
		_, err = relocateFile(options.SingleFile, imageMapper, "", options.Output, baseFlattener)
		return err
	}
	return relocateManifest(options.Manifest, imageMapper, options.Images, options.Output)
}

func createImageMapper(options RelocateImagesOptions) (*imageMapper, error) {
	imageManifest, err := NewImageManifest(options.Images)
	if err != nil {
		return nil, err
	}

	imageMapper, err := newImageMapper(options.Registry, options.RegistryUser, keys(imageManifest.Images), options.Flatten)
	if err != nil {
		return nil, err
	}

	return imageMapper, nil
}

func keys(m map[imageName]imageDigest) []imageName {
	keys := make([]imageName, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func relocateFile(yamlFile string, mapper *imageMapper, base string, outputPath string, strat uriFlattener) (string, error) {
	y, err := fileutils.Read(yamlFile, base)
	if err != nil {
		return "", err
	}

	output := outputFile(outputPath, yamlFile, strat)
	return filepath.Base(output), ioutil.WriteFile(output, mapper.mapImages(y), outputFilePermissions)
}

type uriFlattener func(string) string

func baseFlattener(input string) string {
	return filepath.Base(input)
}

func md5Flattener(input string) string {
	base := baseFlattener(input)
	if base == "manifest.yaml" {
		return base
	}
	hasher := md5.New()
	hasher.Write([]byte(input))
	return base + "-" + hex.EncodeToString(hasher.Sum(nil))
}

func sha256Flattener(input string) string {
	base := baseFlattener(input)
	if base == "manifest.yaml" {
		return base
	}
	hasher := sha256.New()
	hasher.Write([]byte(input))
	return base + "-" + hex.EncodeToString(hasher.Sum(nil))
}

var flatteners = []uriFlattener{baseFlattener, md5Flattener, sha256Flattener}

func relocateManifest(manifestPath string, mapper *imageMapper, imageManifestPath string, outputPath string) error {
	if err := ensureDirectory(outputPath); err != nil {
		return err
	}

	manifest, err := NewManifest(manifestPath)
	if err != nil {
		return err
	}

	nonCollidingFlattener := findNonCollidingFlattener(manifest)

	if nonCollidingFlattener == nil {
		return errors.New("cannot relocate manifest due to collisions in output paths")
	}

	manifestDir, err := fileutils.Dir(manifestPath)
	if err != nil {
		return err
	}

	outputManifest := &Manifest{
		ManifestVersion: manifest.ManifestVersion,
	}

	outputManifest.Istio, err = relocateYamls(manifest.Istio, manifestDir, mapper, outputPath, nonCollidingFlattener)
	if err != nil {
		return err
	}
	outputManifest.Knative, err = relocateYamls(manifest.Knative, manifestDir, mapper, outputPath, nonCollidingFlattener)
	if err != nil {
		return err
	}
	outputManifest.Namespace, err = relocateYamls(manifest.Namespace, manifestDir, mapper, outputPath, nonCollidingFlattener)
	if err != nil {
		return err
	}

	outputManifestPath := filepath.Join(outputPath, nonCollidingFlattener("./manifest.yaml"))
	outputManifestBytes, err := yaml.Marshal(&outputManifest)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(outputManifestPath, outputManifestBytes, outputFilePermissions)
	if err != nil {
		return err
	}

	err = relocateImageManifest(imageManifestPath, mapper, outputPath)
	if err != nil {
		return err
	}

	return copyImages(filepath.Dir(imageManifestPath), outputPath)
}

func copyImages(inputDir string, outputDir string) error {
	imagesPath := filepath.Join(inputDir, "images")

	// if there are no binary images, do not attempt to copy them
	if !isDirectory(imagesPath) {
		return nil
	}

	cmd := exec.Command("cp", "-r", imagesPath, outputDir)
	return cmd.Run()
}

func relocateImageManifest(imageManifestPath string, mapper *imageMapper, outputPath string) error {
	imageManifest, err := NewImageManifest(imageManifestPath)
	if err != nil {
		return err
	}

	relocatedImages := make(map[imageName]imageDigest)
	for n, d := range imageManifest.Images {
		relocatedImages[applyMapper(n, mapper)] = d
	}

	relocatedImageManifest := ImageManifest{
		ManifestVersion: imageManifestVersion_0_1,
		Images:          relocatedImages,
	}

	outputImageManifestPath := filepath.Join(outputPath, "image-manifest.yaml")
	outputImageManifestBytes, err := yaml.Marshal(&relocatedImageManifest)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(outputImageManifestPath, outputImageManifestBytes, outputFilePermissions)
}

func applyMapper(name imageName, mapper *imageMapper) imageName {
	quotedName := imageName(mapper.mapImages([]byte(fmt.Sprintf("%q", name))))
	return quotedName[1 : len(quotedName)-1]
}

func findNonCollidingFlattener(manifest *Manifest) uriFlattener {
	for _, f := range flatteners {
		if collisionless(manifest, f) {
			return f
		}
	}
	return nil
}

func collisionless(manifest *Manifest, f uriFlattener) bool {
	// check that no new fields have been added to the manifest to avoid maintainability issues
	if reflect.Indirect(reflect.ValueOf(manifest)).Type().NumField() != 4 {
		panic("unexpected number of fields in manifest")
	}

	unflattened := make(map[string]string) // maps flattened name to original, unflattened name

	// flatten all the input file names and check for collisions. The manifest filename is special-cased since,
	// regardless of the actual file name of the input manifest, the output file name of the manifest is always
	// "manifest.yaml" and this is always flattened to "manifest.yaml", regardless of the flattener used.
	for _, array := range [][]string{manifest.Istio, manifest.Knative, manifest.Namespace, {"./manifest.yaml"}} {
		for _, input := range array {
			output := f(input)
			if collidingInput, ok := unflattened[output]; ok && input != collidingInput {
				fmt.Printf("Warning: collision between %s and %s. Trying another flattening strategy.\n", collidingInput, input)
				return false
			}
			unflattened[output] = input
		}
	}
	return true
}

func relocateYamls(yamls []string, manifestDir string, mapper *imageMapper, outputPath string, flattener uriFlattener) ([]string, error) {
	rel := []string{}
	for _, y := range yamls {
		yRel, err := relocateFile(y, mapper, manifestDir, outputPath, flattener)
		if err != nil {
			return []string{}, err
		}
		rel = append(rel, yRel)
	}
	return rel, nil
}

func outputFile(outputPath string, yamlFile string, flattener uriFlattener) string {
	if isDirectory(outputPath) {
		return filepath.Join(outputPath, flattener(yamlFile))
	}
	return outputPath
}

func isDirectory(path string) bool {
	f, err := os.Stat(path)
	return err == nil && f.IsDir()
}

func ensureDirectory(path string) error {
	f, err := os.Stat(path)
	if err == nil && f.IsDir() {
		return nil
	}
	if err == nil {
		return fmt.Errorf("output directory is a file: %s", path)
	}

	if os.IsNotExist(err) {
		return os.MkdirAll(path, outputDirPermissions)
	}
	return err
}
