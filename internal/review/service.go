package review

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

type Store interface {
	Find(ctx context.Context, id string) (*Review, error)
	Save(ctx context.Context, review *Review, expectedRevision int) error
}

type MemoryStore struct {
	mu      sync.RWMutex
	reviews map[string]*Review
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{reviews: make(map[string]*Review)}
}

func (store *MemoryStore) Find(ctx context.Context, id string) (*Review, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	store.mu.RLock()
	defer store.mu.RUnlock()
	review, exists := store.reviews[id]
	if !exists {
		return nil, fmt.Errorf("review %s: %w", id, ErrReviewMissing)
	}
	return review.Clone(), nil
}

func (store *MemoryStore) Save(ctx context.Context, review *Review, expectedRevision int) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if review == nil {
		return ErrReviewMissing
	}
	store.mu.Lock()
	defer store.mu.Unlock()
	current, exists := store.reviews[review.ID]
	if exists && current.Revision != expectedRevision {
		return ErrReviewConflict
	}
	store.reviews[review.ID] = review.Clone()
	return nil
}

type Service struct {
	store Store
	now   func() time.Time
}

func NewService(store Store, now func() time.Time) *Service {
	if now == nil {
		now = time.Now
	}
	return &Service{store: store, now: now}
}

func (service *Service) Open(ctx context.Context, review *Review) error {
	if review == nil {
		return ErrReviewMissing
	}
	return service.store.Save(ctx, review.Clone(), 0)
}

func (service *Service) Vote(ctx context.Context, id string, vote Vote) (*Review, error) {
	review, err := service.store.Find(ctx, id)
	if err != nil {
		return nil, err
	}
	expectedRevision := review.Revision
	if err := review.AddVote(vote, service.now()); err != nil {
		return nil, err
	}
	if err := service.store.Save(ctx, review, expectedRevision); err != nil {
		return nil, err
	}
	return review.Clone(), nil
}

func (service *Service) Await(ctx context.Context, id string, interval time.Duration) (*Review, error) {
	if interval <= 0 {
		interval = 10 * time.Millisecond
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		review, err := service.store.Find(ctx, id)
		if err != nil {
			return nil, err
		}
		if review.Decision != DecisionPending {
			return review.Clone(), nil
		}
		if service.now().After(review.Deadline) {
			return nil, ErrReviewExpired
		}
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("await review decision: %w", ctx.Err())
		case <-ticker.C:
		}
	}
}

func IsTerminal(err error) bool {
	return errors.Is(err, ErrReviewExpired) || errors.Is(err, ErrReviewConflict)
}
