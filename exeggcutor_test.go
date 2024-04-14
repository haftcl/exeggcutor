package exeggcutor

import (
	"context"
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

func (r *RunnerImpl) Execute(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		time.Sleep(time.Duration(100) * time.Millisecond)
		x++
		return nil
	}
}

func (h *HydratorImpl) Get(ctx context.Context) ([]Runner, error) {
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

func TestStartStop(t *testing.T) {
	x = 0
	h := &HydratorImpl{}

	// Create a new executor
	e := New(h, 1000, nil, 1)

	go func() {
		time.Sleep(time.Duration(10) * time.Millisecond)
		signalChan <- os.Interrupt
	}()

	e.Start(context.Background())
}

func TestRun(t *testing.T) {
	x = 0
	h := &HydratorImpl{}

	// Create a new executor
	e := New(h, 1000, nil, 5)
	e.runTasks(context.Background())

	if x != 10 {
		t.Errorf("Expected x to be 10, got %d", x)
	}
}

func TestRunConcurrent(t *testing.T) {
	x = 0
	h := &HydratorImpl{}

	// Create a new executor
	e := New(h, 1000, nil, 5)
	e.runTasks(context.Background())

	if x != 10 {
		t.Errorf("Expected x to be 10, got %d", x)
	}
}

type BadRunnerImpl struct {
}

func (r *BadRunnerImpl) Execute(ctx context.Context) error {
	return fmt.Errorf("Bad runner")
}

type BadHydratorImpl struct {
}

func (h *BadHydratorImpl) Get(ctx context.Context) ([]Runner, error) {
	runners := make([]Runner, 0)
	runners = append(runners, &BadRunnerImpl{})

	return runners, nil
}

func TestTaskWithError(t *testing.T) {
	x = 0
	h := &BadHydratorImpl{}

	// Create a new executor
	e := New(h, 1000, nil, 5)
	e.ExitOnError = true

	err := e.runTasks(context.Background())

	if err == nil {
		t.Errorf("Expected error")
	}
}

type HydratorCancelledImpl struct {
}

type RunnerCancelledImpl struct {
	Id int
}

func (r *RunnerCancelledImpl) Execute(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		time.Sleep(time.Duration(500) * time.Millisecond)
		x++

		if x == 5 {
			return fmt.Errorf("Error on 5th runner")
		}

		return nil
	}
}

func (h *HydratorCancelledImpl) Get(ctx context.Context) ([]Runner, error) {
	runners := make([]Runner, 0)

	runners = append(runners, &RunnerCancelledImpl{Id: 1})
	runners = append(runners, &RunnerCancelledImpl{Id: 2})
	runners = append(runners, &RunnerCancelledImpl{Id: 3})
	runners = append(runners, &RunnerCancelledImpl{Id: 4})
	runners = append(runners, &RunnerCancelledImpl{Id: 5})
	runners = append(runners, &RunnerCancelledImpl{Id: 6})
	runners = append(runners, &RunnerCancelledImpl{Id: 7})
	runners = append(runners, &RunnerCancelledImpl{Id: 8})
	runners = append(runners, &RunnerCancelledImpl{Id: 9})
	runners = append(runners, &RunnerCancelledImpl{Id: 10})

	return runners, nil
}

func TestTaskWithCancel(t *testing.T) {
	x = 0
	h := &HydratorCancelledImpl{}

	// Create a new executor
	e := New(h, 1000, nil, 5)
	e.ExitOnError = true
	err := e.runTasks(context.Background())

	if err == nil {
		t.Errorf("Expected error")
	}

	if x == 5 {
		t.Errorf("Expected to run 5 tasks before cancel, ran %d", x)
	}
}
