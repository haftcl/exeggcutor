package exeggcutor

import (
	"context"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"
)

type safeCounter struct {
	mu      sync.Mutex
	counter int
}

func (c *safeCounter) Inc() {
	c.mu.Lock()
	c.counter++
	c.mu.Unlock()
}

func (c *safeCounter) Count() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.counter
}

func (c *safeCounter) Reset() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.counter = 0
}

var counter = safeCounter{counter: 0}

type HydratorImpl struct {
}

type RunnerImpl struct {
	ID int
}

func (r *RunnerImpl) Execute(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		time.Sleep(time.Duration(100) * time.Millisecond)
		counter.Inc()
		return nil
	}
}

func (h *HydratorImpl) Get(ctx context.Context) ([]Runner, error) {
	runners := make([]Runner, 0)

	runners = append(runners, &RunnerImpl{ID: 1})
	runners = append(runners, &RunnerImpl{ID: 2})
	runners = append(runners, &RunnerImpl{ID: 3})
	runners = append(runners, &RunnerImpl{ID: 4})
	runners = append(runners, &RunnerImpl{ID: 5})
	runners = append(runners, &RunnerImpl{ID: 6})
	runners = append(runners, &RunnerImpl{ID: 7})
	runners = append(runners, &RunnerImpl{ID: 8})
	runners = append(runners, &RunnerImpl{ID: 9})
	runners = append(runners, &RunnerImpl{ID: 10})

	return runners, nil
}

func TestStartStop(t *testing.T) {
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
	counter.Reset()
	h := &HydratorImpl{}

	// Create a new executor
	e := New(h, 1000, nil, 5)
	err := e.runTasks(context.Background())

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if counter.Count() != 10 {
		t.Errorf("Expected x to be 10, got %d", counter.Count())
	}
}

func TestRunConcurrent(t *testing.T) {
	counter.Reset()
	h := &HydratorImpl{}

	// Create a new executor
	e := New(h, 1000, nil, 5)
	err := e.runTasks(context.Background())

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if counter.Count() != 10 {
		t.Errorf("Expected x to be 10, got %d", counter.Count())
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
	counter.Reset()
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
	ID int
}

func (r *RunnerCancelledImpl) Execute(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		time.Sleep(time.Duration(100) * time.Millisecond)

		if counter.Count() == 5 {
			return fmt.Errorf("Error on 5th runner")
		}

		counter.Inc()

		return nil
	}
}

func (h *HydratorCancelledImpl) Get(ctx context.Context) ([]Runner, error) {
	runners := make([]Runner, 0)

	runners = append(runners, &RunnerCancelledImpl{ID: 1})
	runners = append(runners, &RunnerCancelledImpl{ID: 2})
	runners = append(runners, &RunnerCancelledImpl{ID: 3})
	runners = append(runners, &RunnerCancelledImpl{ID: 4})
	runners = append(runners, &RunnerCancelledImpl{ID: 5})
	runners = append(runners, &RunnerCancelledImpl{ID: 6})
	runners = append(runners, &RunnerCancelledImpl{ID: 7})
	runners = append(runners, &RunnerCancelledImpl{ID: 8})
	runners = append(runners, &RunnerCancelledImpl{ID: 9})
	runners = append(runners, &RunnerCancelledImpl{ID: 10})

	return runners, nil
}

func TestTaskWithCancel(t *testing.T) {
	counter.Reset()
	h := &HydratorCancelledImpl{}

	// Create a new executor
	e := New(h, 1000, nil, 1)
	e.ExitOnError = true
	err := e.runTasks(context.Background())

	if err == nil {
		t.Errorf("Expected error")
	}

	if counter.Count() != 5 {
		t.Errorf("Expected to run 5 tasks before cancel, ran %d", counter.Count())
	}
}
