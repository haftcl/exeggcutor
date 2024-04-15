# Exeggcutor

Lightweight, no dependency, concurrent executor for running tasks.

## Disclaimer

This is a work in progress and is not yet ready for production use. Its mostly used to learn about concurrency in Golang.

## Usage

```go
package main

import (
    "time"
    "github.com/haftcl/exeggcutor"
)

// The hydrator interface is used to hydrate the executor with tasks that implements the Runner interface
type HydratorImpl struct {
}

// The runner interface is used to define the task to be executed
type RunnerImpl struct {
}

// Implement the interfaces
func (r *RunnerImpl) Execute(ctx context.Context) error {
    time.Sleep(time.Duration(100) * time.Millisecond)
    x++
    return nil
}

func (h *HydratorImpl) Get(ctx context.Context) ([]Runner, error) {
	runners := make([]Runner, 0)
	runners = append(runners, &RunnerImpl{})
	return runners, nil
}

func main() {
    // Create a new hydrator
    h := HydratorImpl{}
    // Create a new exeggcutor
    e := exeggcutor.New(h, 1000, nil, 1)
    // Run the executor
    e.Start(context.Background())
}
```
