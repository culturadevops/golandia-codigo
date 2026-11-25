package main

import (
	"fmt"
	"sync"
	"errors"
)

type TaskError struct {
	msg string
	err error
}

func (te *TaskError) Error() string {
	return te.msg
}

func (te *TaskError) Unwrap() error {
	return te.err
}

func processTask(taskID int, result chan<- error) {
	// Simulamos una tarea que puede fallar
	if taskID%3 == 0 {
		result <- fmt.Errorf("task %d failed", taskID)
	} else {
		result <- nil
	}
}

func main() {
	var wg sync.WaitGroup
	var taskCounter int
	var taskMutex sync.Mutex
	var errorsCount int
	var taskResult chan error

	taskResult = make(chan error, 10)

	for i := 1; i <= 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			err := <-taskResult
			if err != nil {
				taskMutex.Lock()
				errorsCount++
				taskMutex.Unlock()
			} else {
				taskCounter++
			}
		}(i)
	}

	for i := 1; i <= 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			processTask(i, taskResult)
		}()
	}

	wg.Wait()
	close(taskResult)

	fmt.Printf("Total tasks processed: %d\n", taskCounter)
	if errors, ok := <-taskResult; ok {
		fmt.Printf("An error occurred: %v\n", errors)
	} else {
		fmt.Println("All tasks completed successfully.")
	}
}
