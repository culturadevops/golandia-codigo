package main

import (
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

type Task struct {
	name  string
	fcn   func() error
}

var taskCount int64

type TaskError struct {
	Err error
}

func (e *TaskError) Error() string {
	return e.Err.Error()
}

func (e *TaskError) Unwrap() error {
	return e.Err
}

func (e *TaskError) AddWrap(err error) {
	if e.Err == nil {
		e.Err = err
	} else {
		e.Err = errors.Newfmt.Errorf("%w: %w", e.Err, err)
	}
}

func processTask(t Task, wg *sync.WaitGroup, done chan<- error) {
	defer wg.Done()
	if err := t.fcn(); err != nil {
		atomic.AddInt64(&taskCount, -1)
		done <- &TaskError{Err: err}
	} else {
		atomic.AddInt64(&taskCount, 1)
	}
}

func main() {
	tasks := []Task{
		{"Task 1", func() error { time.Sleep(1 * time.Second); return nil }},
		{"Task 2", func() error { time.Sleep(2 * time.Second); return errors.New("Task 2 failed") }},
		{"Task 3", func() error { time.Sleep(3 * time.Second); return nil }},
	}

	var wg sync.WaitGroup
	done := make(chan *TaskError, len(tasks))

	for _, task := range tasks {
		wg.Add(1)
		go processTask(task, &wg, done)
	}

	go func() {
		wg.Wait()
		close(done)
	}()

	var failedTasks []error
	for err := range done {
		failedTasks = append(failedTasks, err)
	}

	fmt.Printf("Tasks completed: %d, Failed: %d\n", atomic.LoadInt64(&taskCount), len(failedTasks))
	if len(failedTasks) > 0 {
		for _, err := range failedTasks {
			fmt.Printf("Failed task: %s, Error: %v\n", err.(*TaskError).Err)
		}
	}
}
