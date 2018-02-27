package retry

import (
	"time"

	"github.com/juju/errgo"
)

// Do performs the given operation. Based on the options, it can retry the operation,
// if it failed.
//
// The following options are supported:
// * RetryChecker(func(err error) bool) - If this func returns true for the returned error, the operation is tried again
// * MaxTries(int) - Maximum number of calls to op() before aborting with MaxRetriesReached
// * Timeout(time.Duration) - Maximum number of time to try to perform this op before aborting with TimeoutReached
// * Sleep(time.Duration) - time to sleep after error failed op()
//
// Defaults:
//  Timeout = 15 seconds
//  MaxRetries = 5
//  Retryer = errgo.Any
//  Sleep = No sleep
//
func Do(op func() error, retryOptions ...RetryOption) error {
	options := newRetryOptions(retryOptions...)

	var timeout <-chan time.Time
	if options.Timeout > 0 {
		timeout = time.After(options.Timeout)
	}

	tryCounter := 0
	for {
		// Check if we reached the timeout
		select {
		case <-timeout:
			return errgo.Mask(TimeoutError, errgo.Any)
		default:
		}

		// Execute the op
		tryCounter++
		lastError := op()
		options.AfterRetry(lastError)

		if lastError != nil {
			if options.Checker != nil && options.Checker(lastError) {
				// Check max retries
				if tryCounter >= options.MaxTries {
					options.AfterRetryLimit(lastError)
					return errgo.WithCausef(lastError, MaxRetriesReachedError, "retry limit reached (%d/%d)", tryCounter, options.MaxTries)
				}

				if options.Sleep > 0 {
					time.Sleep(options.Sleep)
				}
				continue
			}

			return errgo.Mask(lastError, errgo.Any)
		}
		return nil
	}
}
