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
	"encoding/hex"
	"fmt"
	"path"
	"strings"

	"github.com/docker/distribution/reference"
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

func newImageMapper(mappedHost string, mappedUser string, images []imageName, flatten bool) (*imageMapper, error) {
	if err := containsAny(mappedHost, "/", `"`, " "); err != nil {
		return nil, fmt.Errorf("invalid registry hostname: %v", err)
	}
	if err := containsAny(mappedUser, "/", `"`, " "); err != nil {
		return nil, fmt.Errorf("invalid user: %v", err)
	}

	var pathMapping func(string, string) string
	if flatten {
		pathMapping = flattenRepoPath
	} else {
		pathMapping = sanitiseRepoPath
	}

	mapImg := mapImage(pathMapping)

	replacements := []string{}
	for _, img := range images {
		imgHost, imgUser, imgRepoPath, err := parseParts(img)
		if err != nil {
			return nil, err
		}

		fullImg := path.Join(imgHost, imgUser, imgRepoPath)
		mapped := mapImg(mappedHost, mappedUser, imgRepoPath, fullImg)

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

func newIdentityImageMapper() *imageMapper {
	return &imageMapper{
		replacer: strings.NewReplacer(),
	}
}

func mapImage(pathMapping func(string, string) string) func(mappedHost string, mappedUser string, imgRepoPath string, originalImage string) string {
	return func(mappedHost string, mappedUser string, imgRepoPath string, originalImage string) string {
		// ensure local-only images are tagged (with something other than latest) to prevent docker
		// later attempting to download from dev.local (which doesn't really exist)
		mappedPath := pathMapping(imgRepoPath, originalImage)
		if mappedHost == "dev.local" && !strings.Contains(mappedPath, ":") {
			mappedPath = mappedPath + ":local"
		}
		return path.Join(mappedHost, mappedUser, mappedPath)
	}
}

func sanitiseRepoPath(repoPath string, _ string) string {
	// Replace "@sha256:" since digests are not allowed as part of a docker tag
	return strings.Replace(repoPath, "@sha256:", "-", 1)
}

func flattenRepoPath(repoPath string, originalImage string) string {
	hasher := md5.New()
	hasher.Write([]byte(originalImage))
	return fmt.Sprintf("%s-%s", imageBaseName(repoPath), hex.EncodeToString(hasher.Sum(nil)))
}

func imageBaseName(repoPath string) string {
	var base string
	digestPathComponents := strings.SplitN(repoPath, "@sha256:", 2)
	if len(digestPathComponents) == 2 {
		base = path.Base(digestPathComponents[0])
	} else {
		taggedPathComponents := strings.SplitN(repoPath, ":", 2)
		if len(taggedPathComponents) == 2 {
			base = path.Base(taggedPathComponents[0])
		} else {
			base = path.Base(repoPath)
		}
	}
	return base
}

func containsAny(s string, items ...string) error {
	for _, i := range items {
		if strings.Contains(s, i) {
			return fmt.Errorf("'%s' contains '%s'", s, i)
		}
	}
	return nil
}

func parseParts(img imageName) (host string, user string, name string, err error) {
	ref, err := reference.ParseNormalizedNamed(img.String()) // TODO: move this to original source of img
	if err != nil {
		return "", "", "", fmt.Errorf("invalid image: %v", err)
	}

	s := strings.SplitN(ref.Name(), "/", 3)
	switch len(s) {
	case 0:
		panic("SplitN produced empty array")
	case 1:
		panic(fmt.Sprintf("invalid image: domain missing: %s", img))
	case 2:
		return "", "", "", fmt.Errorf("invalid image: user missing: %s", img)
	default:
		return s[0], s[1], s[2], nil
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
