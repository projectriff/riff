package commands

import "sync"

func ApplyInParallel(names []string, deleteResource func(string) error) []error {
	count := len(names)
	deleteErrors := make([]error, count)
	waitGroup := sync.WaitGroup{}
	waitGroup.Add(count)
	for i, name := range names {
		go func(i int, name string) {
			deleteErrors[i] = deleteResource(name)
			waitGroup.Done()
		}(i, name)
	}
	waitGroup.Wait()
	return deleteErrors
}