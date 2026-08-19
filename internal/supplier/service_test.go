package supplier

import (
	"errors"
	"testing"
	"time"
)

type failingGateway struct{ err error }

func (gateway failingGateway) Fetch(request *Request) ([]byte, error) { return nil, gateway.err }

func TestHSupplierRetrySemantics(t *testing.T) {
	now := time.Date(2026, 8, 19, 12, 0, 0, 0, time.UTC)
	store := NewMemoryStore()
	request, err := NewRequest("REQ-1", "SUP-2", "LOT-S", now)
	if err != nil {
		t.Fatal(err)
	}
	service := NewService(failingGateway{err: ErrEvidencePending}, store, func() time.Time { return now.Add(time.Minute) })
	if err := service.RequestEvidence(request); err != nil {
		t.Fatal(err)
	}
	_, err = service.Retrieve(request.ID)
	if !errors.Is(err, ErrEvidencePending) || Classify(err) != FailureRetryable || !ShouldRetry(err, 1, 3) {
		t.Fatalf("pending evidence lost retry semantics: %v", err)
	}
	stored, err := store.Find(request.ID)
	if err != nil {
		t.Fatal(err)
	}
	if stored.Status != StatusPending || stored.Attempts != 1 {
		t.Fatalf("pending state was not persisted: %#v", stored)
	}
	_, err = store.Find("missing")
	if !errors.Is(err, ErrEvidenceMissing) || Classify(err) != FailurePermanent || ShouldRetry(err, 0, 3) {
		t.Fatalf("missing request became retryable: %v", err)
	}
	var missing *Request
	if err := missing.MarkPending(now); !errors.Is(err, ErrEvidenceMissing) {
		t.Fatalf("nil request lost missing evidence identity: %v", err)
	}
	blocked := request.Clone()
	blocked.Status = StatusRejected
	if err := blocked.MarkPending(now); !errors.Is(err, ErrSupplierBlocked) {
		t.Fatalf("blocked supplier lost permanent identity: %v", err)
	}
}
