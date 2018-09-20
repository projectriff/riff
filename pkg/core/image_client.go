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
 *
 */

package core

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/projectriff/riff/pkg/docker"
)

type ImageClient interface {
	LoadAndTagImages(options LoadAndTagImagesOptions) error
	PushImages(options PushImagesOptions) error
	PullImages(options PullImagesOptions) error
}

type PushImagesOptions struct {
	Images string
}

type LoadAndTagImagesOptions struct {
	Images string
}

type PullImagesOptions struct {
	Images             string
	Output             string
	ContinueOnMismatch bool
}

type imageClient struct {
	docker docker.Docker
}

func (c *imageClient) LoadAndTagImages(options LoadAndTagImagesOptions) error {
	_, err := c.loadAndTagImages(options.Images)
	return err
}

func (c *imageClient) PushImages(options PushImagesOptions) error {
	imManifest, err := c.loadAndTagImages(options.Images)
	if err != nil {
		return err
	}
	for name, _ := range imManifest.Images {
		if err := c.docker.PushImage(string(name)); err != nil {
			return err
		}
	}
	return nil
}

func (c *imageClient) loadAndTagImages(imageManifest string) (*ImageManifest, error) {
	imManifest, err := NewImageManifest(imageManifest)
	if err != nil {
		return nil, err
	}
	distroLocation := filepath.Dir(imageManifest)
	for name, digest := range imManifest.Images {
		if digest == "" {
			return nil, fmt.Errorf("image manifest %s does not specify a digest for image %s", imageManifest, name)
		}
		filename := filepath.Join(distroLocation, "images", string(digest))
		if err := c.docker.LoadAndTagImage(string(name), string(digest), filename); err != nil {
			return nil, err
		}
	}
	return imManifest, nil
}

func (c *imageClient) PullImages(options PullImagesOptions) error {
	originalManifest, err := NewImageManifest(options.Images)
	if err != nil {
		return err
	}
	var imagesDir string
	var newManifestPath string
	if options.Output == "" {
		newManifestPath = options.Images
		imagesDir = filepath.Join(filepath.Dir(options.Images), "images")
	} else {
		newManifestPath = filepath.Join(options.Output, "image-manifest.yaml")
		imagesDir = filepath.Join(options.Output, "images")
	}
	if _, err := os.Stat(imagesDir); err != nil {
		if err2 := os.MkdirAll(imagesDir, outputDirPermissions); err2 != nil {
			return err2
		}
	}

	newManifest := EmptyImageManifest()

	for name, sha := range originalManifest.Images {
		if newSha, err := c.docker.PullImage(string(name), imagesDir); err != nil {
			return err
		} else if newSha != string(sha) && sha != "" && !options.ContinueOnMismatch {
			return fmt.Errorf("image %q had digest %v in the original manifest, but the pulled version has digest %s", name, sha, newSha)
		} else {
			newManifest.Images[name] = imageDigest(newSha)
		}
	}

	return newManifest.save(newManifestPath)
}

func NewImageClient(docker docker.Docker) ImageClient {
	return &imageClient{docker: docker}
}
