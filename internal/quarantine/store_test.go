package quarantine

import (
	"errors"
	"testing"
	"time"
)

func TestBQuarantineErrorClassification(t *testing.T) {
	now := time.Date(2026, 8, 19, 10, 0, 0, 0, time.UTC)
	store := NewMemoryStore()
	service := NewService(store, func() time.Time { return now })
	_, err := service.Close("missing", "engineer")
	if !errors.Is(err, ErrHoldMissing) || Classify(err) != FailureMissing || ShouldRetry(err) {
		t.Fatalf("missing hold lost classification: %v", err)
	}
	hold, err := NewHold("H-8", "LOT-8", "particle excursion", "owner-a", now)
	if err != nil {
		t.Fatal(err)
	}
	if err := service.Create(hold); err != nil {
		t.Fatal(err)
	}
	_, err = service.Close(hold.ID, "owner-b")
	if !errors.Is(err, ErrHoldConflict) || Classify(err) != FailureConflict || ShouldRetry(err) {
		t.Fatalf("ownership conflict lost classification: %v", err)
	}
	stale := hold.Clone()
	stale.Version++
	err = store.Put(stale, 99)
	if !errors.Is(err, ErrHoldConflict) || ShouldRetry(err) {
		t.Fatalf("persistence conflict became retryable: %v", err)
	}
}
