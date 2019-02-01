package commands

import "sync"

type CorrelatedResult struct {
	Input string
	Error error
}

func ApplyInParallel(names []string, action func(string) error) []CorrelatedResult {
	count := len(names)
	results := make([]CorrelatedResult, count)
	waitGroup := sync.WaitGroup{}
	waitGroup.Add(count)
	for i, name := range names {
		go func(i int, name string) {
			results[i] = CorrelatedResult{
				Input: name,
				Error: action(name),
			}
			waitGroup.Done()
		}(i, name)
	}
	waitGroup.Wait()
	return results
}
