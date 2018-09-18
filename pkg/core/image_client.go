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
	PushImages(options PushImagesOptions) error
	PullImages(options PullImagesOptions) error
}

type PushImagesOptions struct {
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

// TODO: provide unit test
func (c *imageClient) PushImages(options PushImagesOptions) error {
	imManifest, err := NewImageManifest(options.Images)
	if err != nil {
		return err
	}
	distroLocation := filepath.Dir(options.Images)
	for name, digest := range imManifest.Images {
		filename := filepath.Join(distroLocation, "images", string(digest))
		if err := c.docker.PushImage(string(name), string(digest), filename); err != nil {
			return err
		}
	}
	return nil
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
	if _, err := os.Stat(imagesDir); err != nil && os.IsNotExist(err) {
		if err2 := os.MkdirAll(imagesDir, outputDirPermissions); err2 != nil {
			return err2
		}
	}

	newManifest := EmptyImageManifest()

	for name, sha := range originalManifest.Images {
		if newSha, err := c.docker.PullImage(string(name), imagesDir); err != nil {
			return err
		} else if newSha != string(sha) && sha != "" && !options.ContinueOnMismatch {
			return fmt.Errorf("image %q had digest %v in the original manifest, but the pulled version now has digest %s", name, sha, newSha)
		} else {
			newManifest.Images[name] = imageDigest(newSha)
		}
	}

	return newManifest.save(newManifestPath)
}

func NewImageClient(docker docker.Docker) ImageClient {
	return &imageClient{docker: docker}
}
