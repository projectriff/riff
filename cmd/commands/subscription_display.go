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

package commands

import (
	"fmt"
	"io"

	"github.com/knative/eventing/pkg/apis/channels/v1alpha1"
)

type stringExtractor func(*v1alpha1.Subscription) string

type namedExtractor struct {
	name string
	fn   stringExtractor
}

type subscriptionTable struct {
	height  int
	widths  []int
	content [][]string
}

func display(out io.Writer, subscriptions *[]v1alpha1.Subscription) {
	if len(*subscriptions) == 0 {
		fmt.Fprintln(out, "No resources found")
		return
	}
	subscriptionDisplay := makeSubscriptionDisplay(subscriptions)
	for j := 0; j < subscriptionDisplay.height; j++ {
		for i, width := range subscriptionDisplay.widths {
			fmt.Fprintf(out, "%-*s", width, subscriptionDisplay.content[i][j])
		}
		fmt.Fprintln(out)
	}
}

func makeSubscriptionDisplay(subscriptions *[]v1alpha1.Subscription) *subscriptionTable {
	extractors := makeExtractors()
	widths := make([]int, len(*extractors))
	height := 1 + len(*subscriptions)
	content := make2dArray(len(*extractors), height)
	for i, extractor := range *extractors {
		width := len(extractor.name)
		content[i][0] = extractor.name
		for j, subscription := range *subscriptions {
			value := extractor.fn(&subscription)
			content[i][j+1] = value
			width = max(width, len(value))
		}
		widths[i] = 1 + width
	}
	return &subscriptionTable{
		height:  height,
		widths:  widths,
		content: content,
	}
}

func makeExtractors() *[]namedExtractor {
	return &[]namedExtractor{
		{
			name: "NAME",
			fn:   func(s *v1alpha1.Subscription) string { return s.Name },
		},
		{
			name: "CHANNEL",
			fn:   func(s *v1alpha1.Subscription) string { return s.Spec.Channel },
		},
		{
			name: "SUBSCRIBER",
			fn:   func(s *v1alpha1.Subscription) string { return s.Spec.Subscriber },
		},
		{
			name: "REPLY-TO",
			fn:   func(s *v1alpha1.Subscription) string { return s.Spec.ReplyTo },
		},
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
