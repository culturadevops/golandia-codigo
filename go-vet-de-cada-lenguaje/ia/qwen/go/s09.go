package main

import (
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

type Task struct {
	name   string
	dur    time.Duration
	failure error
}

func (t *Task) Run() error {
	time.Sleep(t.dur)
	if t.failure != nil {
		return fmt.Errorf("task %s failed: %w", t.name, t.failure)
	}
	return nil
}

func processTask(t *Task, done chan *Task, errc chan error, mu *sync.Mutex, count *int32) {
	if err := t.Run(); err != nil {
		errc <- err
		return
	}
	mu.Lock()
	*count += 1
	mu.Unlock()
	done <- t
}

func main() {
	tasks := []Task{
		{"Task1", time.Second, nil},
		{"Task2", 2 * time.Second, errors.New("task-specific error")},
		{"Task3", 3 * time.Second, nil},
	}

	done := make(chan *Task)
	errc := make(chan error)
	mu := new(sync.Mutex)
	count := int32(0)

	for _, task := range tasks {
		go processTask(&task, done, errc, mu, &count)
	}

	for i := 0; i < len(tasks); i++ {
		t := <-done
		fmt.Printf("Task %s completed\n", t.name)
	}

	for i := 0; i < len(tasks); i++ {
		if err := <-errc; err != nil {
			fmt.Printf("Error: %v\n", err)
		}
	}

	fmt.Printf("Total tasks completed: %d\n", atomic.LoadInt32(&count))
}
