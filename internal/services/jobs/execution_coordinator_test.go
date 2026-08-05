package jobs

import (
	"context"
	"errors"
	"testing"
)

func TestExecutionCoordinatorRejectsDuplicateAndReleasesKey(t *testing.T) {
	coordinator := NewExecutionCoordinator(t.Context())
	entered, release := make(chan struct{}), make(chan struct{})
	if err := coordinator.Start(t.Context(), "job", func(context.Context) error {
		close(entered)
		<-release
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	<-entered
	if err := coordinator.Run(t.Context(), "job", func(context.Context) error { return nil }); !errors.Is(err, ErrExecutionAlreadyRunning) {
		t.Fatalf("duplicate error = %v", err)
	}
	close(release)
	coordinator.Stop()
}

func TestExecutionCoordinatorStopCancelsAndWaits(t *testing.T) {
	coordinator := NewExecutionCoordinator(t.Context())
	entered, exited := make(chan struct{}), make(chan struct{})
	if err := coordinator.Start(t.Context(), "job", func(ctx context.Context) error {
		close(entered)
		<-ctx.Done()
		close(exited)
		return ctx.Err()
	}); err != nil {
		t.Fatal(err)
	}
	<-entered
	coordinator.Stop()
	select {
	case <-exited:
	default:
		t.Fatal("Stop returned before the execution exited")
	}
	if err := coordinator.Start(t.Context(), "other", func(context.Context) error { return nil }); !errors.Is(err, ErrExecutionStopped) {
		t.Fatalf("start after stop error = %v", err)
	}
}
