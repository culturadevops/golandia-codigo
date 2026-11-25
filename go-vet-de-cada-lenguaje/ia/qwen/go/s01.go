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

func (t *Task) Run() error {
	return t.f()
}

type TaskError struct {
	err error
}

func (te *TaskError) Error() string {
	return te.err.Error()
}

func (te *TaskError) Unwrap() error {
	return te.err
}

func processTask(task *Task, done chan bool, mutex *sync.Mutex, failed *TaskError) {
	if err := task.Run(); err != nil {
		*failed = TaskError{err: err}
	}
	done <- true
}

func main() {
	tasks := []Task{
		{func() error { return fmt.Errorf("task1 error") }},
		{func() error { return nil }},
		{func() error { return errors.New("task3 error") }},
	}

	var failed TaskError
	var done chan bool = make(chan bool, len(tasks))
	var mutex sync.Mutex
	var completed int32

	for _, task := range tasks {
		go processTask(&task, done, &mutex, &failed)
	}

	for i := 0; i < len(tasks); i++ {
		<-done
		if failed.err != nil {
			atomic.AddInt32(&completed, 1)
		}
	}

	if failed.err != nil {
		fmt.Printf("Failed task: %v\n", failed)
	} else {
		fmt.Printf("All tasks completed successfully. Completed tasks: %d\n", atomic.LoadInt32(&completed))
	}
}
