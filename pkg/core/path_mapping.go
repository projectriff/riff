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

	"github.com/projectriff/riff/pkg/image"
)

type pathMapping func(string, string, image.Name) string

func flattenRepoPath(repoPath string, user string, originalImage image.Name) string {
	hasher := md5.New()
	hasher.Write([]byte(originalImage.String()))
	return fmt.Sprintf("%s/%s-%s", user, path.Base(repoPath), hex.EncodeToString(hasher.Sum(nil)))
}
