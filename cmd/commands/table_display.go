package commands

import (
	"fmt"
	"io"
)

type NamedExtractor struct {
	name string
	fn   StringExtractor
}

type StringExtractor func(interface{}) string

type table struct {
	height  int
	widths  []int
	content [][]string
}

func Display(out io.Writer, items []interface{}, extractors []NamedExtractor) {
	if len(items) == 0 {
		fmt.Fprintln(out, "No resources found")
		return
	}
	display := makeDisplay(items, extractors)
	for j := 0; j < display.height; j++ {
		for i, width := range display.widths {
			fmt.Fprintf(out, "%-*s", width, display.content[i][j])
		}
		fmt.Fprintln(out)
	}
}

func makeDisplay(items []interface{}, extractors []NamedExtractor) *table {
	widths := make([]int, len(extractors))
	height := 1 + len(items)
	content := make2dArray(len(extractors), height)
	for i, extractor := range extractors {
		width := len(extractor.name)
		content[i][0] = extractor.name
		for j, item := range items {
			value := extractor.fn(item)
			content[i][j+1] = value
			width = max(width, len(value))
		}
		widths[i] = 1 + width
	}
	return &table{
		height:  height,
		widths:  widths,
		content: content,
	}
}

func make2dArray(width int, height int) [][]string {
	content := make([][]string, width)
	for i := range content {
		content[i] = make([]string, height)
	}
	return content
}

func max(a int, b int) int {
	if a > b {
		return a
	}
	return b
}
