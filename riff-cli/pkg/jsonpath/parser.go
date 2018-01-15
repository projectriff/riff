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
