package exeggcutor

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golang.org/x/sync/errgroup"
)

// Runner is an interface that defines the methods to execute a job
type Runner interface {
	Execute(context.Context) error
}

// Hydrator is an interface that defines the methods to obtain the jobs
type Hydrator interface {
	Get(context.Context) ([]Runner, error)
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
		logger = slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		}))
		logger = logger.With("app", "exeggcutor")
	}

	return &executor{
		hydrator:    hydrator,
		timeOutInMs: timeOutInMs,
		ExitOnError: false,
		log:         logger,
		numWorkers:  numWorkers,
	}
}

func (e *executor) runTasks(ctx context.Context) error {
	runners, err := e.hydrator.Get(ctx)

	if err != nil {
		e.log.Error("Error getting runners: ", err)
		return err
	}

	group, ctx := errgroup.WithContext(ctx)
	group.SetLimit(e.numWorkers)
	e.log.Info("Running new Tasks", slog.Int("num_tasks", len(runners)))

	for _, runner := range runners {
		runner := runner
		group.Go(func() error {
			e.log.Debug("Executing runner", slog.Any("runner", runner))
			run_error := runner.Execute(ctx)

			if run_error != nil {
				e.log.Error("Error executing runner", "error", run_error)
				return run_error
			}

			e.log.Debug("Runner executed", slog.Any("runner", runner))
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

func (e *executor) Start(ctx context.Context) {
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	for {
		select {
		case <-signalChan:
			e.log.Info("Received signal to stop")
			return
		default:
			err := e.runTasks(ctx)

			if err != nil && e.ExitOnError {
				e.log.Info("Exiting due to error")
				return
			}
		}
	}
}
