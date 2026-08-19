package quarantine

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

type Store interface {
	Find(id string) (*Hold, error)
	Put(hold *Hold, expectedVersion int) error
}

type MemoryStore struct {
	mu    sync.RWMutex
	holds map[string]*Hold
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{holds: make(map[string]*Hold)}
}

func (store *MemoryStore) Find(id string) (*Hold, error) {
	store.mu.RLock()
	defer store.mu.RUnlock()
	hold, ok := store.holds[id]
	if !ok {
		return nil, fmt.Errorf("query hold %q: %w", id, ErrHoldMissing)
	}
	return hold.Clone(), nil
}

func (store *MemoryStore) Put(hold *Hold, expectedVersion int) error {
	if hold == nil {
		return ErrHoldMissing
	}
	store.mu.Lock()
	defer store.mu.Unlock()
	current, exists := store.holds[hold.ID]
	if exists && current.Version != expectedVersion {
		return fmt.Errorf("persist hold %q: %w", hold.ID, ErrHoldConflict)
	}
	store.holds[hold.ID] = hold.Clone()
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

func (service *Service) Create(hold *Hold) error {
	if hold == nil {
		return ErrHoldMissing
	}
	return service.store.Put(hold.Clone(), 0)
}

func (service *Service) Close(id, actor string) (*Hold, error) {
	hold, err := service.store.Find(id)
	if err != nil {
		return nil, fmt.Errorf("load hold for close: %w", err)
	}
	if err := ValidateClose(hold, actor); err != nil {
		return nil, fmt.Errorf("validate close: %w", err)
	}
	expectedVersion := hold.Version
	if err := hold.Close(service.now()); err != nil {
		return nil, fmt.Errorf("close hold: %w", err)
	}
	if err := service.store.Put(hold, expectedVersion); err != nil {
		return nil, fmt.Errorf("save closed hold: %w", err)
	}
	return hold.Clone(), nil
}

func ShouldRetry(err error) bool {
	return err != nil && !errors.Is(err, ErrHoldMissing) && !errors.Is(err, ErrHoldConflict)
}
