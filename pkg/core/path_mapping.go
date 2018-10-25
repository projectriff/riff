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
	"strings"

	"github.com/docker/distribution/reference"

	"github.com/projectriff/riff/pkg/image"
)

type pathMapping func(host string, user string, originalImage image.Name) image.Name

func flattenRepoPath(host string, user string, originalImage image.Name) image.Name {
	hasher := md5.New()
	hasher.Write([]byte(originalImage.String()))
	hash := hex.EncodeToString(hasher.Sum(nil))
	available := reference.NameTotalLengthMax - len(mappedPath(host, user, "", hash))
	fp := flatPath(originalImage.Path(), available)
	var mp string
	if fp == "" {
		mp = fmt.Sprintf("%s/%s/%s", host, user, hash)
	} else {
		mp = mappedPath(host, user, fp, hash)
	}
	mn, err := image.NewName(mp)
	if err != nil {
		panic(err) // handle more gracefully
	}
	return mn
}

func mappedPath(host string, user string, repoPath string, hash string) string {
	return fmt.Sprintf("%s/%s/%s-%s", host, user, repoPath, hash)
}

func flatPath(repoPath string, size int) string {
	return strings.Join(crunch(strings.Split(repoPath, "/"), size), "-")
}

func crunch(components []string, size int) []string {
	for n := len(components); n > 0; n-- {
		comp := reduce(components, n)
		if len(strings.Join(comp, "-")) <= size {
			return comp
		}

	}
	return []string{}
}

func reduce(components []string, n int) []string {
	if len(components) < 2 || len(components) <= n {
		return components
	}

	tmp := make([]string, len(components))
	copy(tmp, components)

	last := components[len(tmp)-1]
	if n < 2 {
		return []string{last}
	}

	front := tmp[0 : n-1]
	return append(front, "-", last)
}
