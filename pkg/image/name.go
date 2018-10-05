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

package image

import (
	"path"
	"strings"

	"github.com/docker/distribution/reference"
)

const (
	dockerHubHost     = "docker.io"
	fullDockerHubHost = "index.docker.io"
)

// Name is a named image reference
type Name struct {
	ref reference.Named
}

var EmptyName Name

func init() {
	EmptyName = Name{nil}
}

func NewName(i string) (Name, error) {
	ref, err := reference.ParseNormalizedNamed(i)
	return Name{ref}, err
}

func (img Name) Normalize() Name {
	if img.ref == nil {
		return EmptyName
	}
	ref, err := NewName(img.String())
	if err != nil {
		panic(err) // should never happen
	}
	return ref
}

func (img Name) String() string {
	if img.ref == nil {
		return ""
	}
	return img.ref.String()
}

func (img Name) Path() string {
	_, p := img.parseHostPath()
	return p
}

func (img Name) Tag() string {
	if taggedRef, ok := img.ref.(reference.Tagged); ok {
		return taggedRef.Tag()
	}
	return ""
}

func (img Name) Digest() Digest {
	if digestedRef, ok := img.ref.(reference.Digested); ok {
		return NewDigest(string(digestedRef.Digest()))
	}
	return EmptyDigest
}

// Synonyms returns the equivalent image names for a given image name. The synonyms are not necessarily
// normalized: in particular they may not have a host name.
func (img Name) Synonyms() []Name {
	if img.ref == nil {
		return []Name{EmptyName}
	}
	imgHost, imgRepoPath := img.parseHostPath()
	nameMap := map[Name]struct{}{img: {}}

	if imgHost == dockerHubHost {
		elidedImg := imgRepoPath
		name, err := synonym(img, elidedImg)
		if err == nil {
			nameMap[name] = struct{}{}
		}

		elidedImgElements := strings.Split(elidedImg, "/")
		if len(elidedImgElements) == 2 && elidedImgElements[0] == "library" {
			name, err := synonym(img, elidedImgElements[1])
			if err == nil {
				nameMap[name] = struct{}{}
			}
		}

		fullImg := path.Join(fullDockerHubHost, imgRepoPath)
		name, err = synonym(img, fullImg)
		if err == nil {
			nameMap[name] = struct{}{}
		}

		dockerImg := path.Join(dockerHubHost, imgRepoPath)
		name, err = synonym(img, dockerImg)
		if err == nil {
			nameMap[name] = struct{}{}
		}
	}

	names := []Name{}
	for n, _ := range nameMap {
		names = append(names, n)
	}

	return names
}

func synonym(original Name, newName string) (Name, error) {
	named, err := reference.WithName(newName)
	if err != nil {
		return EmptyName, err
	}

	if taggedRef, ok := original.ref.(reference.Tagged); ok {
		named, err = reference.WithTag(named, taggedRef.Tag())
		if err != nil {
			return EmptyName, err
		}
	}

	if digestedRef, ok := original.ref.(reference.Digested); ok {
		named, err = reference.WithDigest(named, digestedRef.Digest())
		if err != nil {
			return EmptyName, err
		}
	}

	return Name{named}, nil
}

func (img Name) parseHostPath() (host string, repoPath string) {
	s := strings.SplitN(img.ref.Name(), "/", 2)
	if len(s) == 1 {
		return img.Normalize().parseHostPath()
	}
	return s[0], s[1]
}
