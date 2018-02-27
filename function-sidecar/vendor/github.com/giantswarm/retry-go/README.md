retry-go
========

Small helper library to retry operations automatically on certain errors.

## Usage

The `retry` package provides a `Do()` function which can be used to execute a provided function
until it succeds. 

```
op := func() error {
	// Do something that can fail and should be retried here
	return httpClient.CreateUserOnRemoteServer()
}
retry.Do(op, 
         retry.RetryChecker(IsNetOpErr),
         retry.Timeout(15 * time.Second))
```

Besides the `op` itself, you can provide a few options:

 * RetryChecker(func(err error) bool) - If this func returns true for the returned error, the operation is tried again (default: nil - no retries)
 * MaxTries(int) - Maximum number of calls to op() before aborting with MaxRetriesReachedErr
 * Timeout(time.Duration) - Maximum number of time to try to perform this op before aborting with TimeoutReachedErr
 * Sleep(time.Duration) - time to sleep after every failed op()
