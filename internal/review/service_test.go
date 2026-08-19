package review

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestReviewCancellationAndDeadlinesRemainRequestScoped(t *testing.T) {
	now := time.Date(2026, 8, 19, 13, 0, 0, 0, time.UTC)
	store := NewMemoryStore()
	service := NewService(store, func() time.Time { return now })
	review, err := New("REV-1", "LOT-R", 2, now, now.Add(time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	if err := service.Open(context.Background(), review); err != nil {
		t.Fatal(err)
	}
	canceled, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := store.Find(canceled, review.ID); !errors.Is(err, context.Canceled) {
		t.Fatalf("store ignored canceled request: %v", err)
	}
	if _, err := service.Vote(canceled, review.ID, Vote{Engineer: "eng-a", Decision: DecisionApprove, Comment: "data clean", VotedAt: now}); !errors.Is(err, context.Canceled) {
		t.Fatalf("vote ignored canceled request: %v", err)
	}
	started := time.Now()
	if _, err := service.Await(canceled, review.ID, time.Millisecond); !errors.Is(err, context.Canceled) {
		t.Fatalf("await ignored cancellation: %v", err)
	}
	if time.Since(started) > 100*time.Millisecond {
		t.Fatal("canceled review waited for polling interval")
	}
	stored, err := store.Find(context.Background(), review.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(stored.Votes) != 0 || stored.Revision != 1 {
		t.Fatalf("canceled vote polluted review: %#v", stored)
	}
}
