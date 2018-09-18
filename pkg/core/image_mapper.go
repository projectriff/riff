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
	"fmt"
	"path"
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
	replacer *strings.Replacer
}

func newImageMapper(mappedHost string, mappedUser string, images []imageName) (*imageMapper, error) {
	if err := containsAny(mappedHost, "/", `"`, " "); err != nil {
		return nil, fmt.Errorf("invalid registry hostname: %v", err)
	}
	if err := containsAny(mappedUser, "/", `"`, " "); err != nil {
		return nil, fmt.Errorf("invalid user: %v", err)
	}

	replacements := []string{}
	for _, img := range images {
		imgHost, imgUser, imgRepoPath, err := img.parseParts()
		if err != nil {
			return nil, err
		}

		fullImg := path.Join(imgHost, imgUser, imgRepoPath)
		mapped := mapImage(mappedHost, mappedUser, imgRepoPath)

		replacements = append(replacements, quote(fullImg), quote(mapped))
		replacements = append(replacements, spacePrefix(fullImg), spacePrefix(mapped))
		if imgHost == dockerHubHost {
			elidedImg := path.Join(imgUser, imgRepoPath)
			replacements = append(replacements, quote(elidedImg), quote(mapped))
			replacements = append(replacements, spacePrefix(elidedImg), spacePrefix(mapped))

			fullImg := path.Join(fullDockerHubHost, imgUser, imgRepoPath)
			replacements = append(replacements, quote(fullImg), quote(mapped))
			replacements = append(replacements, spacePrefix(fullImg), spacePrefix(mapped))
		}
	}

	return &imageMapper{
		replacer: strings.NewReplacer(replacements...),
	}, nil
}

func mapImage(mappedHost string, mappedUser string, imgRepoPath string) string {
	return path.Join(mappedHost, mappedUser, sanitiseRepoPath(imgRepoPath))
}

func sanitiseRepoPath(repoPath string) string {
	// Replace "@sha256:" since digests are not allowed as part of a docker tag
	return strings.Replace(repoPath, "@sha256:", "-", 1)
}

func containsAny(s string, items ...string) error {
	for _, i := range items {
		if strings.Contains(s, i) {
			return fmt.Errorf("'%s' contains '%s'", s, i)
		}
	}
	return nil
}

func (img imageName) parseParts() (host string, user string, name string, err error) {
	if err := containsAny(string(img), `"`, " "); err != nil {
		return "", "", "", fmt.Errorf("invalid image: %v", err)
	}

	s := strings.SplitN(string(img), "/", 3)
	switch len(s) {
	case 0:
		panic("SplitN produced empty array")
	case 1:
		return "", "", "", fmt.Errorf("invalid image: user missing: %s", img)
	case 2:
		return dockerHubHost, s[0], s[1], nil
	default:
		// Normalise docker hub hosts to a single form
		host = s[0]
		if host == fullDockerHubHost {
			host = dockerHubHost
		}

		return host, s[1], s[2], nil
	}
}

func quote(image string) string {
	return fmt.Sprintf("%q", image)
}

func spacePrefix(image string) string {
	return fmt.Sprintf(" %s", image)
}

func (m *imageMapper) mapImages(input []byte) []byte {
	return []byte(m.replacer.Replace(string(input)))
}
