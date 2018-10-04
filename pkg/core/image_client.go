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
	"io"
	"os"
	"path/filepath"

	"github.com/projectriff/riff/pkg/fileutils"

	"github.com/projectriff/riff/pkg/docker"
)

type ImageClient interface {
	LoadAndTagImages(options LoadAndTagImagesOptions) error
	PushImages(options PushImagesOptions) error
	PullImages(options PullImagesOptions) error
	RelocateImages(options RelocateImagesOptions) error
	DownloadSystem(options DownloadSystemOptions) error
	ListImages(options ListImagesOptions) error
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

type ListImagesOptions struct {
	Manifest string
	Images   string
	Force    bool
	NoCheck  bool
}

type imageLister func(resource string, baseDir string) ([]string, error)

type imageClient struct {
	docker     docker.Docker
	copier     fileutils.Copier
	checker    fileutils.Checker
	listImages imageLister
	log        io.Writer
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

	return newManifest.Save(newManifestPath)
}

func (c *imageClient) ListImages(options ListImagesOptions) error {
	m, err := NewManifest(options.Manifest)
	if err != nil {
		return err
	}
	baseDir := filepath.Dir(options.Manifest)
	imPath := options.Images
	if imPath == "" {
		imPath = filepath.Join(baseDir, "image-manifest.yaml")
	}
	if !options.Force && c.checker.Exists(imPath) {
		return fmt.Errorf("image manifest already exists, use `--force` to overwrite it")
	}

	images := []string{}

	// scan the resources for potential images
	err = m.VisitResources(func(res string) error {
		i, err := c.listImages(res, baseDir)
		if err != nil {
			return err
		}
		images = append(images, i...)
		return nil
	})
	if err != nil {
		return err
	}

	// populate the image manifest, removing duplicates, and, if checking is enabled, warning about images which appear to be invalid
	im := EmptyImageManifest()
	for _, i := range images {
		if _, ok := im.Images[imageName(i)]; ok {
			continue
		}
		if !options.NoCheck {
			fmt.Printf("Checking image %s\n", i)
			if !c.docker.ImageExists(i) {
				fmt.Printf("Warning: omitting image %s which is not known to docker. To include it, re-run with --no-check.\n", i)
				continue
			}
		}
		im.Images[imageName(i)] = ""
	}

	return im.Save(imPath)
}

func NewImageClient(docker docker.Docker, copier fileutils.Copier, checker fileutils.Checker, listImages imageLister, log io.Writer) ImageClient {
	return &imageClient{docker: docker, copier: copier, checker: checker, listImages: listImages, log: log}
}

func (c *imageClient) printf(format string, a ...interface{}) (n int, err error) {
	return fmt.Fprintf(c.log, format, a...)
}
