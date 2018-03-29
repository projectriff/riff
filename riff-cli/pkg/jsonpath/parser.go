/*
 * Copyright 2018 the original author or authors.
 *
 *   Licensed under the Apache License, Version 2.0 (the "License");
 *   you may not use this file except in compliance with the License.
 *   You may obtain a copy of the License at
 *  
 *        http://www.apache.org/licenses/LICENSE-2.0
 *  
 *   Unless required by applicable law or agreed to in writing, software
 *   distributed under the License is distributed on an "AS IS" BASIS,
 *   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *   See the License for the specific language governing permissions and
 *   limitations under the License.
 */

package jsonpath

import (
	"github.com/oliveagle/jsonpath"
	"encoding/json"
)

type Parser struct {
	data interface{}
}

func NewParser(b []byte) *Parser {
	p := Parser{}
	json.Unmarshal(b, &p.data)
	return &p
}

func (p Parser) Value(path string) (interface{}, error) {
	comp, err := jsonpath.Compile(path)
	if err != nil {
		return nil, err
	}
	return comp.Lookup(p.data)
}

func (p Parser) StringValue(path string) (string, error) {
	res, err := p.Value(path)
	return res.(string), err
}
