package main

import (
	"fmt"
	"sync"
	"sync/atomic"
	"errors"
)

type Task struct {
	f func() error
}

var taskCounter uint32

type TaskError struct {
	msg string
	err error
}

func (e *TaskError) Error() string {
	return fmt.Sprintf("Error en tarea: %s", e.msg)
}

func (e *TaskError) Unwrap() error {
	return e.err
}

func main() {
	tasks := []Task{
		{func() error { return nil }},
		{func() error { return fmt.Errorf("error 1") }},
		{func() error { return fmt.Errorf("error 2") }},
	}

	var wg sync.WaitGroup
	var counter uint32
	var mu sync.Mutex

	for _, task := range tasks {
		wg.Add(1)
		go func(task Task) {
			defer wg.Done()
			if err := task.f(); err != nil {
				mu.Lock()
				taskCounter++
				err = &TaskError{msg: "Task failed", err: err}
				mu.Unlock()
				fmt.Printf("Task failed with error: %v\n", err)
			} else {
				mu.Lock()
				atomic.AddUint32(&counter, 1)
				mu.Unlock()
				fmt.Printf("Task completed successfully.\n")
			}
		}(task)
	}

	wg.Wait()
	fmt.Printf("Total tasks completed: %d\n", counter)
}
