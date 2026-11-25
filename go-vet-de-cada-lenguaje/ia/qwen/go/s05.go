package main

import (
	"fmt"
	"sync"
	"errors"
	"sync/atomic"
)

type Task func() error

func main() {
	tasks := []Task{
		func() error { return fmt.Errorf("task 1 failed") },
		func() error { return nil },
		func() error { return fmt.Errorf("task 3 failed") },
	}

	var (
		mu         sync.Mutex
		errored    int32
		processed int
	)

	err := processTasksInParallel(tasks, func(err error) {
		mu.Lock()
		if err != nil {
			atomic.AddInt32(&errored, 1)
		}
		atomic.AddInt32(&processed, 1)
		mu.Unlock()
	})

	if err != nil {
		fmt.Printf("Error processing tasks: %v\n", err)
	} else {
		fmt.Printf("Processed %d tasks, %d failed\n", processed, errored)
	}
}

func processTasksInParallel(tasks []Task, handle func(error)) error {
	var wg sync.WaitGroup
	var mu sync.Mutex
	var err error

	for _, task := range tasks {
		wg.Add(1)
		go func(task Task) {
			defer wg.Done()
			if err := task(); err != nil {
				handle(fmt.Errorf("%w", err))
			}
		}(task)
	}
	wg.Wait()

	if err1 := recover(); err1 != nil {
		err = fmt.Errorf("unhandled error: %v", err1)
	}
	return err
}
