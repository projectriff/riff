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
	"sort"
	"strings"

	"github.com/projectriff/riff/pkg/image"
)

// imageMapper does substring replacement of images with minimal delimiter checking.
// It does not understand the full (YAML) syntax of the input string.
// It certainly does not understand the semantics of the input string.
type imageMapper struct {
	replacer *strings.Replacer
}

func newImageMapper(mappedHost string, mappedUser string, images []image.Name) (*imageMapper, error) {
	if err := containsAny(mappedHost, "/", `"`, " "); err != nil {
		return nil, fmt.Errorf("invalid registry hostname: %v", err)
	}
	if err := containsAny(mappedUser, "/", `"`, " "); err != nil {
		return nil, fmt.Errorf("invalid user: %v", err)
	}

	// Sort images so that if one has a path which is a prefix of another's path, the image with the longer path
	// appears first. This ensures that the correct string replacement is performed first.
	sort.SliceStable(images, func(i, j int) bool {
		return strings.HasPrefix(images[i].Path(), images[j].Path())
	})

	mapImg := mapImage(flattenRepoPath)

	replacements := []string{}
	for _, img := range images {
		mapped := mapImg(mappedHost, mappedUser, img)

		for _, name := range img.Synonyms() {
			fullImg := name.String()
			replacements = append(replacements, quote(fullImg), quote(mapped.String()))
			replacements = append(replacements, spacePrefix(fullImg), spacePrefix(mapped.String()))
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

func mapImage(pathMapping pathMapping) pathMapping {
	name := func(mappedHost string, mappedUser string, originalImage image.Name) image.Name {
		mappedPath := pathMapping(mappedHost, mappedUser, originalImage)

		newTag := ""
		if tag := originalImage.Tag(); tag != "" {
			// preserve original tag
			newTag = tag
		} else if mappedPath.Host() == "dev.local" {
			// ensure local-only images are tagged (with something other than latest) to prevent docker
			// later attempting to download from dev.local (which doesn't really exist)
			newTag = "local"
		}

		if newTag != "" {
			var err error
			mappedPath, err = mappedPath.WithTag(newTag)
			if err != nil {
				panic(err) // should never occur
			}
		}

		return mappedPath
	}
	return name
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
