package review

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestJReviewRequestScope(t *testing.T) {
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
	working := review.Clone()
	first := Vote{Engineer: "eng-a", Decision: DecisionApprove, Comment: "clean", VotedAt: now}
	if err := working.AddVote(first, now); err != nil || working.Revision != 2 {
		t.Fatalf("valid vote did not advance revision: %#v %v", working, err)
	}
	if err := working.AddVote(first, now); !errors.Is(err, ErrReviewConflict) {
		t.Fatalf("duplicate vote was accepted: %v", err)
	}
	expired, err := New("REV-X", "LOT-R", 1, now, now.Add(time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	if err := expired.AddVote(first, now.Add(2*time.Minute)); !errors.Is(err, ErrReviewExpired) {
		t.Fatalf("expired review accepted a vote: %v", err)
	}
}
