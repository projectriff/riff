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
	"strings"
)

const (
	dockerHubHost     = "docker.io"
	fullDockerHubHost = "index.docker.io"
)

// imageMapper does substring replacement of images with minimal delimiter checking.
// It does not understand the full (YAML) syntax of the input string.
// It certainly does not understand the semantics of the input string.
type imageMapper struct {
	mappings map[string]string
}

func newImageMapper(mappedHost string, mappedUser string, images []string) (*imageMapper, error) {
	if err := contains(mappedHost, "/", `"`, " "); err != nil {
		return nil, fmt.Errorf("Invalid registry hostname: %v", err)
	}
	if err := contains(mappedUser, "/", `"`, " "); err != nil {
		return nil, fmt.Errorf("Invalid user: %v", err)
	}

	mappings := make(map[string]string)
	for _, img := range images {
		imgHost, imgUser, imgRepoPath, err := parseImage(img)
		if err != nil {
			return nil, err
		}

		fullImg := fmt.Sprintf("%s/%s/%s", imgHost, imgUser, imgRepoPath)
		mapped := fmt.Sprintf("%s/%s/%s", mappedHost, mappedUser, imgRepoPath)

		mappings[quote(fullImg)] = quote(mapped)
		mappings[spacePrefix(fullImg)] = spacePrefix(mapped)
		if imgHost == dockerHubHost {
			elidedImg := fmt.Sprintf("%s/%s", imgUser, imgRepoPath)
			mappings[quote(elidedImg)] = quote(mapped)
			mappings[spacePrefix(elidedImg)] = spacePrefix(mapped)

			fullImg := fmt.Sprintf("%s/%s/%s", fullDockerHubHost, imgUser, imgRepoPath)
			mappings[quote(fullImg)] = quote(mapped)
			mappings[spacePrefix(fullImg)] = spacePrefix(mapped)
		}
	}

	return &imageMapper{
		mappings: mappings,
	}, nil
}

func contains(s string, items ...string) error {
	for _, i := range items {
		if strings.Contains(s, i) {
			return fmt.Errorf("'%s' contains '%s'", s, i)
		}
	}
	return nil
}

func parseImage(img string) (string, string, string, error) {
	if err := contains(img, `"`, " "); err != nil {
		return "", "", "", fmt.Errorf("Invalid image: %v", err)
	}

	s := strings.SplitN(img, "/", 3)
	switch len(s) {
	case 0:
		panic("SplitN produced empty array")
	case 1:
		return "", "", "", fmt.Errorf("Invalid image: user missing: %s", img)
	case 2:
		return dockerHubHost, s[0], s[1], nil
	default:
		// Normalise docker hub hosts to a single form
		host := s[0]
		if host == fullDockerHubHost {
			host = dockerHubHost
		}

		return host, s[1], strings.Join(s[2:], "/"), nil
	}
}

func quote(image string) string {
	return fmt.Sprintf("%q", image)
}

func spacePrefix(image string) string {
	return fmt.Sprintf(" %s", image)
}

func (m *imageMapper) mapImages(input []byte) []byte {
	for img, mapped := range m.mappings {
		input = bytes.Replace(input, []byte(img), []byte(mapped), -1)
	}
	return input
}
