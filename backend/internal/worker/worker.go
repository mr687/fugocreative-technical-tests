package worker

import (
	"context"
	"log/slog"
	"time"

	"github.com/mr687/fugocreative-technical-tests/internal/model"
	"github.com/mr687/fugocreative-technical-tests/internal/repository"
	"github.com/mr687/fugocreative-technical-tests/internal/service"
)

const concurrency = 2
const pollInterval = 500 * time.Millisecond

// Worker claims durable jobs and processes them outside the request path.
type Worker struct {
	jobs     repository.AIJobRepository
	provider service.AIProvider
	timeout  time.Duration
	wake     chan struct{}
}

func New(jobs repository.AIJobRepository, provider service.AIProvider, timeout time.Duration) *Worker {
	return &Worker{jobs: jobs, provider: provider, timeout: timeout, wake: make(chan struct{}, 1)}
}

func (w *Worker) Wake() {
	select {
	case w.wake <- struct{}{}:
	default:
	}
}

func (w *Worker) Start(ctx context.Context) error {
	recoverCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := w.jobs.RecoverStale(recoverCtx, w.timeout+time.Minute); err != nil {
		return err
	}
	for range concurrency {
		go w.run(ctx)
	}
	w.Wake()
	return nil
}

func (w *Worker) run(ctx context.Context) {
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()
	for {
		job, err := w.jobs.ClaimNext(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			slog.Error("claim AI job failed", "error", err)
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
			case <-w.wake:
			}
			continue
		}
		if job == nil {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
			case <-w.wake:
			}
			continue
		}
		w.process(ctx, *job)
	}
}

func (w *Worker) process(ctx context.Context, job model.AIJob) {
	processCtx, cancel := context.WithTimeout(ctx, w.timeout)
	result, processErr := w.provider.Process(processCtx, job.Brief)
	cancel()
	if ctx.Err() != nil {
		return
	}

	persistCtx, persistCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer persistCancel()
	if err := w.jobs.Complete(persistCtx, job, result, processErr); err != nil {
		slog.Error("persist AI job result failed", "job_id", job.ID, "error", err)
		return
	}
	if processErr != nil {
		slog.Error("AI job failed", "job_id", job.ID, "reason", "provider processing failed")
		return
	}
	slog.Info("AI job completed", "job_id", job.ID)
}
