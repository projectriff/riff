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

	"github.com/projectriff/riff/pkg/image"
)

// imageMapper does substring replacement of images with minimal delimiter checking.
// It does not understand the full (YAML) syntax of the input string.
// It certainly does not understand the semantics of the input string.
type imageMapper struct {
	replacer *strings.Replacer
}

func newImageMapper(mappedHost string, mappedUser string, images []image.Name, flatten bool) (*imageMapper, error) {
	if err := containsAny(mappedHost, "/", `"`, " "); err != nil {
		return nil, fmt.Errorf("invalid registry hostname: %v", err)
	}
	if err := containsAny(mappedUser, "/", `"`, " "); err != nil {
		return nil, fmt.Errorf("invalid user: %v", err)
	}

	var pathMapping pathMapping
	if flatten {
		pathMapping = flattenRepoPath
	} else {
		pathMapping = sanitiseRepoPath
	}

	mapImg := mapImage(pathMapping)

	replacements := []string{}
	for _, img := range images {
		mapped := mapImg(mappedHost, mappedUser, img.Path(), img)

		for _, name := range img.Synonyms() {
			fullImg := name.String()
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

type pathMapping func(string, string, image.Name) string

func mapImage(pathMapping pathMapping) func(mappedHost string, mappedUser string, imgRepoPath string, originalImage image.Name) string {
	return func(mappedHost string, mappedUser string, imgRepoPath string, originalImage image.Name) string {
		// ensure local-only images are tagged (with something other than latest) to prevent docker
		// later attempting to download from dev.local (which doesn't really exist)
		mappedPath := pathMapping(imgRepoPath, mappedUser, originalImage)
		if mappedHost == "dev.local" && !strings.Contains(mappedPath, ":") {
			mappedPath = mappedPath + ":local"
		}
		return path.Join(mappedHost, mappedPath)
	}
}

func sanitiseRepoPath(repoPath string, user string, originalImage image.Name) string {
	repoPathElements := strings.SplitN(repoPath, "/", 2)
	var mapped string
	switch len(repoPathElements) {
	case 1:
		// prevent collisions
		return flattenRepoPath(repoPath, user, originalImage)
	default:
		mapped = path.Join(user, repoPathElements[1])
		if tag := originalImage.Tag(); tag != "" {
			mapped += ":" + tag
		}
		if digest := originalImage.Digest(); digest != image.EmptyDigest {
			mapped += "@" + digest.String()
		}
	}
	// Replace "@sha256:" since digests are not allowed as part of a docker tag
	return strings.Replace(mapped, "@sha256:", "-", 1)
}

func flattenRepoPath(repoPath string, user string, originalImage image.Name) string {
	hasher := md5.New()
	hasher.Write([]byte(originalImage.String()))
	return fmt.Sprintf("%s/%s-%s", user, imageBaseName(repoPath), hex.EncodeToString(hasher.Sum(nil)))
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

func quote(image string) string {
	return fmt.Sprintf("%q", image)
}

func spacePrefix(image string) string {
	return fmt.Sprintf(" %s", image)
}

func (m *imageMapper) mapImages(input []byte) []byte {
	return []byte(m.replacer.Replace(string(input)))
}
