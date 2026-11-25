package main

import (
	"fmt"
	"sync"
	"errors"
	"sync/atomic"
)

type Task struct {
	id     int
	failed bool
}

func (t *Task) Do() error {
	// Simulación de tarea
	if t.id%3 == 0 {
		return fmt.Errorf("task %d failed", t.id)
	}
	return nil
}

type TaskError struct {
	err error
}

func (e *TaskError) Error() string {
	return e.err.Error()
}

func (e *TaskError) Unwrap() error {
	return e.err
}

func processTask(t *Task, wg *sync.WaitGroup, failedTasks *int32) error {
	defer wg.Done()
	err := t.Do()
	if err != nil {
		atomic.AddInt32(failedTasks, 1)
		return &TaskError{err: err}
	}
	return nil
}

func main() {
	tasks := []Task{
		{1}, {2}, {3}, {4}, {5}, {6}, {7}, {8}, {9}, {10},
	}
	var wg sync.WaitGroup
	var failedTasks int32
	var mu sync.Mutex
	var failedCount int

	for _, task := range tasks {
		wg.Add(1)
		go func(t Task) {
			err := processTask(&t, &wg, &failedTasks)
			if err != nil {
				if err.(*TaskError).Unwrap() != nil {
					mu.Lock()
					failedCount++
					fmt.Printf("Task %d failed: %v\n", t.id, err)
					mu.Unlock()
				}
			}
		}(task)
	}

	wg.Wait()
	mu.Lock()
	fmt.Printf("Total tasks: %d, failed tasks: %d\n", len(tasks), failedCount)
	mu.Unlock()
}
