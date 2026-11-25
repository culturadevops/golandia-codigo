package main

import (
	"fmt"
	"sync"
	"sync/atomic"
	"errors"
)

type Task struct {
	task func() error
}

func (t *Task) Run() error {
	return t.task()
}

type TaskProcessor struct {
	tasks  []Task
	mutex  sync.Mutex
	count  int32
	errors []error
}

func (p *TaskProcessor) Add(task Task) {
	p.mutex.Lock()
	p.tasks = append(p.tasks, task)
	p.mutex.Unlock()
}

func (p *TaskProcessor) Process() error {
	var wg sync.WaitGroup
	p.mutex.Lock()
	for _, task := range p.tasks {
		wg.Add(1)
		go func(t Task) {
			defer wg.Done()
			if err := t.Run(); err != nil {
				p.mutex.Lock()
				p.errors = append(p.errors, err)
				atomic.AddInt32(&p.count, 1)
				p.mutex.Unlock()
			}
		}(task)
	}
	p.mutex.Unlock()
	wg.Wait()
	return p.CombineErrors()
}

func (p *TaskProcessor) CombineErrors() error {
	if p.count > 0 {
		return errors.New("some tasks failed").WithWrap(p.errors...)
	}
	return nil
}

func (e *error) WithWrap(errs ...error) error {
	return fmt.Errorf("%w: %v", e, errs)
}

func main() {
	task1 := Task{task: func() error { return fmt.Errorf("task 1 failed") }}
	task2 := Task{task: func() error { return nil }}
	task3 := Task{task: func() error { return fmt.Errorf("task 3 failed") }}

	processor := &TaskProcessor{}
	processor.Add(task1)
	processor.Add(task2)
	processor.Add(task3)

	if err := processor.Process(); err != nil {
		fmt.Printf("Errors encountered: %v\n", err)
	} else {
		fmt.Println("All tasks processed successfully")
	}
}
