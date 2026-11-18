package worker

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/mr687/fugocreative-technical-tests/internal/model"
)

type jobStoreStub struct {
	completed model.AIResult
	err       error
	called    bool
}

func (s *jobStoreStub) ClaimNext(context.Context) (*model.AIJob, error)   { return nil, nil }
func (s *jobStoreStub) RecoverStale(context.Context, time.Duration) error { return nil }
func (s *jobStoreStub) Complete(_ context.Context, _ model.AIJob, result model.AIResult, err error) error {
	s.called, s.completed, s.err = true, result, err
	return nil
}

type providerStub struct {
	result model.AIResult
	err    error
}

func (p providerStub) Process(context.Context, model.Brief) (model.AIResult, error) {
	return p.result, p.err
}

func TestWorkerPersistsProviderErrorVerbatim(t *testing.T) {
	providerErr := errors.New("provider status: 429 account throttled")
	store := &jobStoreStub{}
	worker := New(store, providerStub{err: providerErr}, time.Second)
	worker.process(context.Background(), model.AIJob{ID: "job", BriefID: "brief"})
	if !store.called || store.err != providerErr || store.err.Error() != providerErr.Error() {
		t.Fatalf("persisted error = %v, want original %q", store.err, providerErr.Error())
	}
}

func TestWorkerPersistsSuccessfulResult(t *testing.T) {
	result := model.AIResult{Summary: "Site", ServiceCategory: "Website", Priority: "Low", PriorityReason: "Small scope", MissingInfo: []string{}}
	store := &jobStoreStub{}
	New(store, providerStub{result: result}, time.Second).process(context.Background(), model.AIJob{ID: "job", BriefID: "brief"})
	if !store.called || store.err != nil || !reflect.DeepEqual(store.completed, result) {
		t.Fatalf("persisted result = %+v, error=%v", store.completed, store.err)
	}
}

func TestWorkerLeavesClaimedJobForRecoveryOnShutdown(t *testing.T) {
	store := &jobStoreStub{}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	New(store, providerStub{result: model.AIResult{Summary: "unused"}}, time.Second).process(ctx, model.AIJob{ID: "job", BriefID: "brief"})
	if store.called {
		t.Fatal("completed job after shutdown cancellation; job should remain recoverable")
	}
}
