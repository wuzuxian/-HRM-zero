// Package scheduler is the server's in-process scheduled-task skeleton
// (服务内嵌定时调度骨架). It is deliberately dependency-free: one goroutine
// per registered Job ticking on a time.Ticker — no cron library, no external
// queue (ADR/spec: 不引 XXL-Job/Redis 队列).
//
// Usage (register in cmd wiring / internal/custom, keep business logic in the
// owning domain package):
//
//	s := scheduler.New()
//	s.Register(scheduler.Job{
//	    Name:       "hrm-transfer-effective-scan",
//	    Interval:   24 * time.Hour,
//	    RunAtStart: true, // catch up on everything missed while the server was down
//	    Run:        func(ctx context.Context) error { return scanPendingEffective(ctx) },
//	})
//	s.Start() // non-blocking; each job gets its own goroutine
//	...
//	s.Stop()  // cancels job contexts, waits for the current runs to finish
//
// Reusable by later tickets (HRM-10 离职生效、假勤日结/月报锁定): register one
// Job per domain scan; errors are logged per run and never kill the loop, so a
// failing scan degrades to "retry next tick".
//
// Note: jobs must be idempotent — RunAtStart + a server restart can run a job
// twice for the same business day (调动生效扫描靠 applied 标记幂等).
package scheduler

import (
	"context"
	"sync"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
)

// Job is one scheduled task registered with the Scheduler.
type Job struct {
	// Name identifies the job in logs. Keep it stable per business scan.
	Name string
	// Interval is how often Run fires. RunAtStart can fire it once earlier.
	Interval time.Duration
	// RunAtStart, when true, runs Run once immediately on Start (before the
	// first tick) — used for daily scans so a restart backfills missed days.
	RunAtStart bool
	// Run is one execution of the task. Returned errors are logged, not fatal.
	Run func(ctx context.Context) error
}

// Scheduler runs registered Jobs in separate goroutines. Zero-value Scheduler
// is not usable; construct with New.
type Scheduler struct {
	mu      sync.Mutex
	jobs    []Job
	cancel  context.CancelFunc
	started bool
	wg      sync.WaitGroup
}

// New returns an empty Scheduler. Register Jobs before Start.
func New() *Scheduler {
	return &Scheduler{}
}

// Register adds a Job. Must be called before Start (after Start it is ignored).
func (s *Scheduler) Register(job Job) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.started {
		logx.Errorf("scheduler: Register(%s) after Start ignored", job.Name)
		return
	}
	s.jobs = append(s.jobs, job)
}

// Start launches one goroutine per registered job. Idempotent-safe: a second
// Start on a running or stopped Scheduler is a logged no-op.
func (s *Scheduler) Start() {
	s.mu.Lock()
	if s.started {
		s.mu.Unlock()
		logx.Error("scheduler: Start called twice")
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	s.cancel = cancel
	s.started = true
	jobs := make([]Job, len(s.jobs))
	copy(jobs, s.jobs)
	s.mu.Unlock()

	for _, job := range jobs {
		s.wg.Add(1)
		go s.runJob(ctx, job)
	}
}

// Stop cancels all job contexts and waits for in-flight runs to finish.
// Safe to call multiple times and on a never-started Scheduler.
func (s *Scheduler) Stop() {
	s.mu.Lock()
	cancel := s.cancel
	s.cancel = nil
	s.mu.Unlock()
	if cancel != nil {
		cancel()
	}
	s.wg.Wait()
}

func (s *Scheduler) runJob(ctx context.Context, job Job) {
	defer s.wg.Done()
	logx.Infof("scheduler: job %s started (interval %s, runAtStart %v)", job.Name, job.Interval, job.RunAtStart)

	if job.RunAtStart {
		runOnce(ctx, job)
	}
	ticker := time.NewTicker(job.Interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			logx.Infof("scheduler: job %s stopped", job.Name)
			return
		case <-ticker.C:
			runOnce(ctx, job)
		}
	}
}

func runOnce(ctx context.Context, job Job) {
	if job.Run == nil {
		return
	}
	runCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	if err := job.Run(runCtx); err != nil {
		logx.Errorf("scheduler: job %s run failed: %v", job.Name, err)
	}
}
