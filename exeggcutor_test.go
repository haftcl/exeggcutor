package exeggcutor

import (
	"fmt"
	"os"
	"testing"
	"time"
)

var x = 0

type HydratorImpl struct {
}

type RunnerImpl struct {
	Id int
}

func (r *RunnerImpl) Execute() error {
	fmt.Println("Executing job", r.Id)
	time.Sleep(time.Duration(1) * time.Second)
	x++
	return nil
}

func (h *HydratorImpl) Get() ([]Runner, error) {
	runners := make([]Runner, 0)

	runners = append(runners, &RunnerImpl{Id: 1})
	runners = append(runners, &RunnerImpl{Id: 2})
	runners = append(runners, &RunnerImpl{Id: 3})
	runners = append(runners, &RunnerImpl{Id: 4})
	runners = append(runners, &RunnerImpl{Id: 5})
	runners = append(runners, &RunnerImpl{Id: 6})
	runners = append(runners, &RunnerImpl{Id: 7})
	runners = append(runners, &RunnerImpl{Id: 8})
	runners = append(runners, &RunnerImpl{Id: 9})
	runners = append(runners, &RunnerImpl{Id: 10})

	return runners, nil
}

func TestStart(t *testing.T) {
	x = 0
	h := &HydratorImpl{}

	// Create a new executor
	e := New(h, 1000, nil, 1)

	go func() {
		time.Sleep(time.Duration(10) * time.Millisecond)
		signalChan <- os.Interrupt
	}()

	e.Start()

	if x != 10 {
		t.Errorf("Expected x to be 10, got %d", x)
	}
}

func TestStartConcurrent(t *testing.T) {
	x = 0
	h := &HydratorImpl{}

	// Create a new executor
	e := New(h, 1000, nil, 5)

	go func() {
		time.Sleep(time.Duration(10) * time.Millisecond)
		signalChan <- os.Interrupt
	}()

	e.Start()

	if x != 10 {
		t.Errorf("Expected x to be 10, got %d", x)
	}
}

type BadRunnerImpl struct {
}

func (r *BadRunnerImpl) Execute() error {
	return fmt.Errorf("Bad runner")
}

type BadHydratorImpl struct {
}

func (h *BadHydratorImpl) Get() ([]Runner, error) {
	runners := make([]Runner, 0)
	runners = append(runners, &BadRunnerImpl{})

	return runners, nil
}

func TestTaskWithError(t *testing.T) {
	h := &BadHydratorImpl{}

	// Create a new executor
	e := New(h, 1000, nil, 5)
	e.ExitOnError = true

	err := e.runTasks()

	if err == nil {
		t.Errorf("Expected error")
	}
}