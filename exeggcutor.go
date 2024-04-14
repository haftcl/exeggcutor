package exeggcutor

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golang.org/x/sync/errgroup"
)

// Runner is an interface that defines the methods to execute a job
type Runner interface {
	Execute() error
}

// Hydrator is an interface that defines the methods to obtain the jobs
type Hydrator interface {
	Get() ([]Runner, error)
}

type executor struct {
	hydrator    Hydrator
	numWorkers  int
	timeOutInMs int
	ExitOnError bool
	log         *slog.Logger
}

func New(hydrator Hydrator, timeOutInMs int, logger *slog.Logger, numWorkers int) *executor {
	if logger == nil {
		logger = slog.New(slog.NewJSONHandler(os.Stderr, nil))
	}

	return &executor{
		hydrator:    hydrator,
		timeOutInMs: timeOutInMs,
		ExitOnError: false,
		log:         logger,
		numWorkers:  numWorkers,
	}
}

func (e *executor) runTasks() error {
	runners, err := e.hydrator.Get()

	if err != nil {
		e.log.Error("Error getting runners: ", err)
		return err
	}

	var group errgroup.Group
	group.SetLimit(e.numWorkers)
	e.log.Info("Running new Tasks", slog.Int("num_tasks", len(runners)))

	for _, runner := range runners {
		runner := runner
		group.Go(func() error {
			run_error := runner.Execute()

			if run_error != nil {
				e.log.Error("Error executing runner: ", run_error)
				return run_error
			}

			return nil
		})
	}

	if err := group.Wait(); err != nil && e.ExitOnError {
		return err
	}

	time.Sleep(time.Duration(e.timeOutInMs) * time.Millisecond)
	return nil
}

var signalChan = make(chan os.Signal, 1)

func (e *executor) Start() {
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	for {
		select {
		case <-signalChan:
			e.log.Info("Received signal to stop")
			return
		default:
			err := e.runTasks()

			if err != nil && e.ExitOnError {
				e.log.Info("Exiting due to error")
				return
			}
		}
	}
}
