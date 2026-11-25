package main

import (
	"fmt"
	"sync"
	"sync/atomic"
	"errors"
	"time"
)

type Task struct {
功能  string
}

func (t *Task) Do() error {
	time.Sleep(100 * time.Millisecond)
	if t.功能 == "Fail" {
		return fmt.Errorf("task failed")
	}
	return nil
}

type TaskError struct {
原错误 error
Wrap   error
}

func (e *TaskError) Error() string {
	if e.Wrap != nil {
		return e.原错误.Error() + ": " + e.Wrap.Error()
	}
	return e.原错误.Error()
}

func (e *TaskError) Unwrap() error {
	return e.Wrap
}

func main() {
	var tasks = []Task{
		{"Task1"},
		{"Task2"},
		{"Task3"},
		{"Fail"},
		{"Task4"},
	}

	var wg sync.WaitGroup
	var counter int32
	var mu sync.Mutex

	for _, task := range tasks {
		wg.Add(1)
		go func(t Task) {
			defer wg.Done()
			err := t.Do()
			if err != nil {
				mu.Lock()
				atomic.AddInt32(&counter, 1)
				mu.Unlock()
				if atomic.LoadInt32(&counter) == 1 {
					innerErr := &TaskError{
						原错误: err,
					}
					err = errors.Wrap(err, "task failed with inner error")
					err = errors.Wrap(err, innerErr)
				}
			}
		}(task)
	}
	wg.Wait()

	fmt.Printf("Tasks completed with %d errors\n", atomic.LoadInt32(&counter))
}
