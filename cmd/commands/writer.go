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

package commands

import (
	"io"

	"os"

	"fmt"

	"github.com/ghodss/yaml"
)

type Marshaller interface {
	Marshal(o interface{}) error
}

type marshaller struct {
	io.Writer
}

func NewMarshaller(name string, force bool) (Marshaller, error) {
	_, err := os.Stat(name)
	if os.IsNotExist(err) || force {
		f, err := os.Create(name)
		if err != nil {
			return nil, err
		}
		return &marshaller{Writer: f}, nil
	}
	return nil, fmt.Errorf("not overwriting existing %q file", name)
}

func (w *marshaller) Marshal(o interface{}) error {
	bs, err := yaml.Marshal(o)
	if err != nil {
		return err
	}
	_, err = w.Write(bs)
	if err != nil {
		return err
	}
	_, err = w.Write([]byte("---\n"))
	return err
}
