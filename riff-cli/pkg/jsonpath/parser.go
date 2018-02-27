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
	nodeprime "github.com/NodePrime/jsonpath"
	"strings"
)

type Parser struct {
	Json []byte
}

func NewParser(json []byte) *Parser {
	return &Parser{Json: json}
}

func (p Parser) Value(path string) string {
	paths, err := nodeprime.ParsePaths(path)
	if err != nil {
		panic(err)
	}
	eval, err := nodeprime.EvalPathsInBytes(p.Json, paths)
	if err != nil {
		panic(err)
	}
	if r, ok := eval.Next(); ok {
		if (r.Type == nodeprime.JsonString) {
			return strings.Replace(string(r.Value),"\"","",-1)
		} else {
			return string(r.Value)
		}
	} else {
		return ""
	}
}
