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

package resource

import (
	"io/ioutil"
	"net/http"
	"net/url"
	"path/filepath"
)

func Load(release string, baseDir string) ([]byte, error) {
	releaseUrl, err := url.Parse(release)
	if err != nil {
		return nil, err
	}

	var content []byte
	if releaseUrl.Scheme == "" {
		content, err = ioutil.ReadFile(filepath.Join(baseDir, releaseUrl.Path))
		if err != nil {
			return nil, err
		}
	} else {
		resp, err := http.Get(releaseUrl.String())
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		content, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
	}
	return content, nil
}
