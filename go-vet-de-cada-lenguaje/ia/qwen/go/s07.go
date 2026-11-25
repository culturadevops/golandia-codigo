package main

import (
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

type Task struct {
	id    int
	fcn   func() error
}

func (t *Task) Run() error {
	return t.fcn()
}

func (t *Task) Error() error {
	return nil
}

func (t *Task) WrapErrorf(format string, args ...interface{}) error {
	return fmt.Errorf("%d: %w", t.id, fmt.Errorf(format, args...))
}

type TaskRunner struct {
	tasks  []Task
	mu     sync.Mutex
	ready  int32
	ended  int32
}

func NewTaskRunner(tasks []Task) *TaskRunner {
	return &TaskRunner{
		tasks: tasks,
	}
}

func (r *TaskRunner) Run() error {
	var wg sync.WaitGroup
	for _, t := range r.tasks {
		wg.Add(1)
		go func(t Task) {
			defer wg.Done()
			if err := t.Run(); err != nil {
				r.mu.Lock()
				defer r.mu.Unlock()
				r.ended++
				err = t.WrapErrorf("Task %d failed: %w", t.id, err)
				fmt.Println(err)
			} else {
				r.mu.Lock()
				defer r.mu.Unlock()
				r.ready++
			}
		}(t)
	}
	wg.Wait()
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.ready+r.ended != len(r.tasks) {
		return errors.New("not all tasks completed")
	}
	return nil
}

func main() {
	tasks := []Task{
		{1, func() error { time.Sleep(200 * time.Millisecond); return nil }},
		{2, func() error { time.Sleep(200 * time.Millisecond); return fmt.Errorf("task 2 error") }},
		{3, func() error { time.Sleep(200 * time.Millisecond); return nil }},
	}

	tr := NewTaskRunner(tasks)
	if err := tr.Run(); err != nil {
		fmt.Println("Error:", err)
	}
}
