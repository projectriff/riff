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
	"path/filepath"

	"github.com/projectriff/riff/pkg/docker"
)

type ImageClient interface {
	PushImages(options PushImagesOptions) error
}

type PushImagesOptions struct {
	Images string
}

type imageClient struct {
	docker docker.Docker
}

func (c *imageClient) PushImages(options PushImagesOptions) error {
	imManifest, err := NewImageManifest(options.Images)
	if err != nil {
		return err
	}
	distroLocation := filepath.Dir(options.Images)
	for name, sha := range imManifest.Images {
		filename := fmt.Sprintf("%s/images/%s", distroLocation, sha)
		if err := c.docker.PushImage(string(name), string(sha), filename); err != nil {
			return err
		}
	}
	return nil
}

func NewImageClient(docker docker.Docker) ImageClient {
	return &imageClient{docker: docker}
}
