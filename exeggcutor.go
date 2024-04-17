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

type Executor struct {
	hydrator    Hydrator
	numWorkers  int
	timeOutInMs int
	log         *slog.Logger
	ExitOnError bool
}

func New(hydrator Hydrator, timeOutInMs int, logger *slog.Logger, numWorkers int) *Executor {
	if logger == nil {
		logger = slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		}))
		logger = logger.With("app", "exeggcutor")
	}

	return &Executor{
		hydrator:    hydrator,
		timeOutInMs: timeOutInMs,
		ExitOnError: false,
		numWorkers:  numWorkers,
		log:         logger,
	}
}

func (e *Executor) runTasks(ctx context.Context) error {
	runners, err := e.hydrator.Get(ctx)

	if err != nil {
		e.log.Error("Error getting runners: ", err)
		return err
	}

	group, ctx := errgroup.WithContext(ctx)
	group.SetLimit(e.numWorkers)
	e.log.Debug("Running new tasks cycle", slog.Int("num_tasks", len(runners)))

	for _, runner := range runners {
		runner := runner
		group.Go(func() error {
			e.log.Debug("Executing runner", slog.Any("runner", runner))
			runError := runner.Execute(ctx)

			if runError != nil {
				e.log.Error("Error executing runner", "error", runError)
				return runError
			}

			e.log.Debug("Runner executed", slog.Any("runner", runner))
			return nil
		})
	}

	err = group.Wait()
	if err != nil && e.ExitOnError {
		e.log.Debug("Tasks returned error, closing executor", slog.Any("error", err))
		return err
	}

	e.log.Debug("Finishing tasks cycle", slog.Int("num_tasks", len(runners)), slog.Any("error", err))
	time.Sleep(time.Duration(e.timeOutInMs) * time.Millisecond)

	return nil
}

var signalChan = make(chan os.Signal, 1)

func (e *Executor) Start(ctx context.Context) {
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	e.log.Info("Starting new executor")

	for {
		select {
		case <-signalChan:
			e.log.Info("Received signal to stop")
			return
		default:
			err := e.runTasks(ctx)

			if err != nil && e.ExitOnError {
				e.log.Debug("Exiting due to error")
				return
			}
		}
	}
}
