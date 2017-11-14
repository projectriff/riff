package retry

import (
	"time"

	"github.com/juju/errgo"
)

const (
	DefaultMaxTries = 3
	DefaultTimeout  = time.Duration(15 * time.Second)
)

// Not is a helper to invert another .
func Not(checker func(err error) bool) func(err error) bool {
	return func(err error) bool {
		return !checker(err)
	}
}

type RetryOption func(options *retryOptions)

// Timeout specifies the maximum time that should be used before aborting the retry loop.
// Note that this does not abort the operation in progress.
func Timeout(d time.Duration) RetryOption {
	return func(options *retryOptions) {
		options.Timeout = d
	}
}

// MaxTries specifies the maximum number of times op will be called by Do().
func MaxTries(tries int) RetryOption {
	return func(options *retryOptions) {
		options.MaxTries = tries
	}
}

// RetryChecker defines whether the given error is an error that can be retried.
func RetryChecker(checker func(err error) bool) RetryOption {
	return func(options *retryOptions) {
		options.Checker = checker
	}
}

func Sleep(d time.Duration) RetryOption {
	return func(options *retryOptions) {
		options.Sleep = d
	}
}

// AfterRetry is called after a retry and can be used e.g. to emit events.
func AfterRetry(afterRetry func(err error)) RetryOption {
	return func(options *retryOptions) {
		options.AfterRetry = afterRetry
	}
}

// AfterRetryLimit is called after a retry limit is reached and can be used
// e.g. to emit events.
func AfterRetryLimit(afterRetryLimit func(err error)) RetryOption {
	return func(options *retryOptions) {
		options.AfterRetryLimit = afterRetryLimit
	}
}

type retryOptions struct {
	Timeout         time.Duration
	MaxTries        int
	Checker         func(err error) bool
	Sleep           time.Duration
	AfterRetry      func(err error)
	AfterRetryLimit func(err error)
}

func newRetryOptions(options ...RetryOption) retryOptions {
	state := retryOptions{
		Timeout:         DefaultTimeout,
		MaxTries:        DefaultMaxTries,
		Checker:         errgo.Any,
		AfterRetry:      func(err error) {},
		AfterRetryLimit: func(err error) {},
	}

	for _, option := range options {
		option(&state)
	}

	return state
}
