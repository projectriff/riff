package test_support

import (
	"net"
	"net/http"
	"time"
)

type HttpResponse struct {
	StatusCode int
	Content    []byte
	Headers    map[string]string
}

func Serve(listener net.Listener, response HttpResponse) error {
	return ServeSlow(listener, response, 0)
}

func ServeSlow(listener net.Listener, response HttpResponse, delay time.Duration) error {
	err := http.Serve(listener, http.HandlerFunc(func(responseWriter http.ResponseWriter, req *http.Request) {
		time.Sleep(delay)
		responseWriter.WriteHeader(statusCodeOrDefault(response, http.StatusOK))
		for headerKey, headerValue := range response.Headers {
			responseWriter.Header().Add(headerKey, headerValue)
		}
		_, _ = responseWriter.Write(response.Content)
	}))
	if err != nil {
		return err
	}
	return nil
}

func statusCodeOrDefault(response HttpResponse, defaultCode int) int {
	code := response.StatusCode
	if code == 0 {
		return defaultCode
	}
	return code
}
