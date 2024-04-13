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
}

func (r *RunnerImpl) Execute() error {
	fmt.Println("Executing job")
	x++
	return nil
}

func (h *HydratorImpl) Get() ([]Runner, error) {
	runners := make([]Runner, 0)
	runners = append(runners, &RunnerImpl{})

	return runners, nil
}

func TestStart(t *testing.T) {
	h := &HydratorImpl{}

	// Create a new executor
	e := New(h, 1000, nil)

	go func() {
		time.Sleep(time.Duration(2) * time.Second)
		signalChan <- os.Interrupt
	}()

	e.Start()

	if x != 3 {
		t.Errorf("Expected x to be 3, got %d", x)
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
	e := New(h, 1000, nil)
	e.ExitOnError = true

	err := e.runTasks()

	if err == nil {
		t.Errorf("Expected error")
	}
}
