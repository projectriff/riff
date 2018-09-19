package commands

import (
	"fmt"
	"github.com/knative/eventing/pkg/apis/channels/v1alpha1"
	"io"
)

type subscriptionDisplay struct {
	headers       *[]subscriptionHeader
	subscriptions *[]v1alpha1.Subscription
}

type stringExtractor func(*v1alpha1.Subscription) string

type subscriptionHeader struct {
	namedExtractor
	padding int
}

type namedExtractor struct {
	name string
	fn   stringExtractor
}

func (display subscriptionDisplay) showHeaders(stdout io.Writer) {
	for _, header := range *(display.headers) {
		fmt.Fprintf(stdout, "%-*s", header.padding, header.name)
	}
	fmt.Fprintln(stdout)
}

func (display subscriptionDisplay) showItems(stdout io.Writer) {
	for _, subscription := range *(display.subscriptions) {
		for _, header := range *(display.headers) {
			fmt.Fprintf(stdout, "%-*s", header.padding, header.fn(&subscription))
		}
		fmt.Fprintln(stdout)
	}
}

func makeSubscriptionDisplay(subscriptions *[]v1alpha1.Subscription) subscriptionDisplay {
	extractors := makeExtractors()
	headers := make([]subscriptionHeader, len(*extractors))
	for i, extractor := range *extractors {
		headers[i] = makeSubscriptionHeader(extractor.name, extractor.fn, subscriptions)
	}
	return subscriptionDisplay{
		headers:       &headers,
		subscriptions: subscriptions,
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

func makeSubscriptionHeader(headerName string, extractor stringExtractor, subscriptions *[]v1alpha1.Subscription) subscriptionHeader {
	result := subscriptionHeader{padding: 1 + max(len(headerName), maxLength(subscriptions, extractor))}
	result.name = headerName
	result.fn = extractor
	return result
}

func maxLength(subscriptions *[]v1alpha1.Subscription, extractor stringExtractor) int {
	result := 0
	for _, subscription := range *subscriptions {
		if s := extractor(&subscription); len(s) > result {
			result = len(s)
		}
	}
	return result
}

func max(a int, b int) int {
	if a > b {
		return a
	}
	return b
}
