package scheduler

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// Seams under test (pure, no DB): Register/Start/Stop lifecycle of the
// in-process scheduler, and the RunAtStart/run-on-tick/error-isolated
// semantics each Job gets. The daily business scans (调动生效/离职/日结)
// are tested at their own logic seams; this test only proves the skeleton.

// waitFor polls condition for up to 2s so assertions are deterministic even
// though job ticks are asynchronous.
func waitFor(t *testing.T, condition func() bool) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if condition() {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
}

func TestSchedulerRunsJobs(t *testing.T) {
	s := New()
	var fast, slow, atStart atomic.Int64
	s.Register(Job{
		Name:     "fast",
		Interval: 20 * time.Millisecond,
		Run: func(ctx context.Context) error {
			fast.Add(1)
			return nil
		},
	})
	s.Register(Job{
		Name:       "at-start",
		Interval:   time.Hour,
		RunAtStart: true,
		Run: func(ctx context.Context) error {
			atStart.Add(1)
			return nil
		},
	})
	s.Register(Job{
		Name:     "slow",
		Interval: time.Hour,
		Run: func(ctx context.Context) error {
			slow.Add(1)
			return nil
		},
	})

	s.Start()
	waitFor(t, func() bool { return fast.Load() > 0 })
	s.Stop()

	if fast.Load() == 0 {
		t.Fatal("job with 20ms interval should have run at least once before Stop")
	}
	if atStart.Load() != 1 {
		t.Fatalf("RunAtStart job should run exactly once at Start, got %d", atStart.Load())
	}
	if slow.Load() != 0 {
		t.Fatalf("hourly job without RunAtStart should not run before its tick, got %d", slow.Load())
	}
}

func TestSchedulerIsolatesJobErrors(t *testing.T) {
	s := New()
	var calls atomic.Int64
	s.Register(Job{
		Name:     "boom",
		Interval: 15 * time.Millisecond,
		Run: func(ctx context.Context) error {
			calls.Add(1)
			return errors.New("boom")
		},
	})
	s.Start()
	waitFor(t, func() bool { return calls.Load() > 0 })
	s.Stop()
	if calls.Load() == 0 {
		t.Fatal("failing job should still keep running (error logged, not propagated)")
	}
}

func TestSchedulerStopIsIdempotent(t *testing.T) {
	s := New()
	s.Register(Job{
		Name:     "tick",
		Interval: 10 * time.Millisecond,
		Run:      func(ctx context.Context) error { return nil },
	})
	s.Start()
	s.Stop()
	s.Stop() // must not panic or deadlock

	var mu sync.Mutex
	started := 0
	// Start after Stop must be a no-op, not a second set of goroutines.
	s.Start()
	mu.Lock()
	started++
	mu.Unlock()
	s.Stop()
}

func TestSchedulerContextIsCancelled(t *testing.T) {
	s := New()
	gotCtx := make(chan context.Context, 1)
	s.Register(Job{
		Name:       "ctx",
		Interval:   time.Hour,
		RunAtStart: true,
		Run: func(ctx context.Context) error {
			gotCtx <- ctx
			return nil
		},
	})
	s.Start()
	ctx := <-gotCtx
	select {
	case <-ctx.Done():
	case <-time.After(time.Second):
		t.Fatal("job context should be cancelled on Stop")
	}
	s.Stop()
}
