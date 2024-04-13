package exeggcutor

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"
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

func New(hydrator Hydrator, timeOutInMs int, logger *slog.Logger) *executor {
	if logger == nil {
		logger = slog.New(slog.NewJSONHandler(os.Stderr, nil))
	}

	return &executor{
		hydrator:    hydrator,
		timeOutInMs: timeOutInMs,
		ExitOnError: false,
		log:         logger,
	}
}

func (e *executor) runTasks() error {
	runners, err := e.hydrator.Get()

	if err != nil {
		e.log.Error("Error getting runners: ", err)
		return err
	}

	for _, runner := range runners {
		run_error := runner.Execute()

		if run_error != nil {
			e.log.Error("Error executing runner: ", run_error)

			if e.ExitOnError {
				return run_error
			}
		}
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
