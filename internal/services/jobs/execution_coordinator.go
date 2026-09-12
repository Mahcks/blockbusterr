package jobs

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"

	"github.com/mahcks/blockbusterr/config"
	"github.com/mahcks/blockbusterr/internal/database"
	"github.com/mahcks/blockbusterr/internal/integrations"
)

var (
	ErrExecutionAlreadyRunning = errors.New("execution is already running")
	ErrExecutionStopped        = errors.New("execution coordinator is stopped")
)

const SelectionCycleExecutionID = "selection-cycle"

type activeExecution struct {
	cancel context.CancelCauseFunc
}

// ExecutionCoordinator owns every live job execution for the application.
type ExecutionCoordinator struct {
	ctx    context.Context
	cancel context.CancelFunc
	mu     sync.Mutex
	active map[string]*activeExecution
	wg     sync.WaitGroup
	closed bool
}

func NewExecutionCoordinator(parent context.Context) *ExecutionCoordinator {
	ctx, cancel := context.WithCancel(parent)
	return &ExecutionCoordinator{ctx: ctx, cancel: cancel, active: make(map[string]*activeExecution)}
}

func (c *ExecutionCoordinator) Run(trigger context.Context, id string, run func(context.Context) error) error {
	execution, runCtx, err := c.acquire(trigger, id)
	if err != nil {
		return err
	}
	defer c.release(id, execution)
	return run(runCtx)
}

func (c *ExecutionCoordinator) Start(trigger context.Context, id string, run func(context.Context) error) error {
	execution, runCtx, err := c.acquire(trigger, id)
	if err != nil {
		return err
	}
	go func() {
		defer c.release(id, execution)
		if err := run(runCtx); err != nil && !errors.Is(err, context.Canceled) {
			slog.Error("job execution failed", "job_id", id, "error", err)
		}
	}()
	return nil
}

func (c *ExecutionCoordinator) RunDynamicJob(ctx context.Context, cfg *config.Config, db *database.Database, job config.DynamicJob, dryRun bool) error {
	return c.Run(ctx, job.ID, func(runCtx context.Context) error {
		return RunDynamicJob(runCtx, cfg, db, job, dryRun)
	})
}

func (c *ExecutionCoordinator) StartDynamicJob(ctx context.Context, cfg *config.Config, db *database.Database, job config.DynamicJob, dryRun bool) error {
	return c.Start(ctx, job.ID, func(runCtx context.Context) error {
		return RunDynamicJob(runCtx, cfg, db, job, dryRun)
	})
}

func (c *ExecutionCoordinator) RunSelectionCycle(ctx context.Context, cfg *config.Config, db *database.Database, dryRun bool) (plan SelectionCyclePlan, err error) {
	err = c.Run(ctx, SelectionCycleExecutionID, func(runCtx context.Context) error {
		plan, err = runSelectionCycle(runCtx, cfg, db, dryRun, func(ctx context.Context, cfg *config.Config, db *database.Database, job config.DynamicJob, dryRun bool, selection map[string]ScoreInfo, movies []integrations.Movie, shows []integrations.Show) (summary JobExecutionSummary, err error) {
			err = c.Run(ctx, job.ID, func(jobCtx context.Context) error {
				summary, err = RunSelectedDynamicJob(jobCtx, cfg, db, job, dryRun, selection, movies, shows)
				return err
			})
			return summary, err
		})
		return err
	})
	return plan, err
}

func (c *ExecutionCoordinator) Stop() {
	c.mu.Lock()
	if !c.closed {
		c.closed = true
		c.cancel()
		for _, execution := range c.active {
			execution.cancel(context.Canceled)
		}
	}
	c.mu.Unlock()
	c.wg.Wait()
}

func (c *ExecutionCoordinator) acquire(trigger context.Context, id string) (*activeExecution, context.Context, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed || c.ctx.Err() != nil || trigger.Err() != nil {
		return nil, nil, ErrExecutionStopped
	}
	if _, exists := c.active[id]; exists {
		return nil, nil, fmt.Errorf("%w: %s", ErrExecutionAlreadyRunning, id)
	}
	runCtx, cancel := context.WithCancelCause(c.ctx)
	stopTrigger := context.AfterFunc(trigger, func() { cancel(context.Cause(trigger)) })
	execution := &activeExecution{cancel: func(cause error) {
		stopTrigger()
		cancel(cause)
	}}
	c.active[id] = execution
	c.wg.Add(1)
	return execution, runCtx, nil
}

func (c *ExecutionCoordinator) release(id string, execution *activeExecution) {
	execution.cancel(nil)
	c.mu.Lock()
	if c.active[id] == execution {
		delete(c.active, id)
	}
	c.mu.Unlock()
	c.wg.Done()
}
