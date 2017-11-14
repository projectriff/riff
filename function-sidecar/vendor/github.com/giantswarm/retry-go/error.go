package retry

import (
	"github.com/juju/errgo"
)

var (
	TimeoutError           = errgo.New("Operation aborted. Timeout occured")
	MaxRetriesReachedError = errgo.New("Operation aborted. Too many errors.")
)

// IsTimeout returns true if the cause of the given error is a TimeoutError.
func IsTimeout(err error) bool {
	return errgo.Cause(err) == TimeoutError
}

// IsMaxRetriesReached returns true if the cause of the given error is a MaxRetriesReachedError.
func IsMaxRetriesReached(err error) bool {
	return errgo.Cause(err) == MaxRetriesReachedError
}
