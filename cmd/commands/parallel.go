package commands

import "sync"

func ApplyInParallel(names []string, action func(string) error) []error {
	count := len(names)
	result := make([]error, count)
	waitGroup := sync.WaitGroup{}
	waitGroup.Add(count)
	for i, name := range names {
		go func(i int, name string) {
			result[i] = action(name)
			waitGroup.Done()
		}(i, name)
	}
	waitGroup.Wait()
	return result
}