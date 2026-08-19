package supplier

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

type Gateway interface {
	Fetch(request *Request) ([]byte, error)
}

type Store interface {
	Save(request *Request) error
	Find(id string) (*Request, error)
}

type MemoryStore struct {
	mu       sync.RWMutex
	requests map[string]*Request
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{requests: make(map[string]*Request)}
}

func (store *MemoryStore) Save(request *Request) error {
	if request == nil {
		return ErrEvidenceMissing
	}
	store.mu.Lock()
	defer store.mu.Unlock()
	store.requests[request.ID] = request.Clone()
	return nil
}

func (store *MemoryStore) Find(id string) (*Request, error) {
	store.mu.RLock()
	defer store.mu.RUnlock()
	request, exists := store.requests[id]
	if !exists {
		return nil, fmt.Errorf("request %s: %v", id, ErrEvidenceMissing)
	}
	return request.Clone(), nil
}

type Service struct {
	gateway Gateway
	store   Store
	now     func() time.Time
}

func NewService(gateway Gateway, store Store, now func() time.Time) *Service {
	if now == nil {
		now = time.Now
	}
	return &Service{gateway: gateway, store: store, now: now}
}

func (service *Service) RequestEvidence(request *Request) error {
	if request == nil {
		return ErrEvidenceMissing
	}
	return service.store.Save(request.Clone())
}

func (service *Service) Retrieve(id string) ([]byte, error) {
	request, err := service.store.Find(id)
	if err != nil {
		return nil, fmt.Errorf("find supplier request: %v", err)
	}
	payload, err := service.gateway.Fetch(request.Clone())
	if err != nil {
		if markErr := request.MarkPending(service.now()); markErr != nil {
			return nil, errors.Join(fmt.Errorf("fetch supplier evidence: %v", err), markErr)
		}
		if saveErr := service.store.Save(request); saveErr != nil {
			return nil, errors.Join(fmt.Errorf("fetch supplier evidence: %v", err), saveErr)
		}
		return nil, fmt.Errorf("fetch supplier evidence: %v", err)
	}
	if err := request.MarkReady(service.now()); err != nil {
		return nil, err
	}
	if err := service.store.Save(request); err != nil {
		return nil, err
	}
	return append([]byte(nil), payload...), nil
}

func ShouldRetry(err error, attempts, maximum int) bool {
	if err == nil || attempts >= maximum {
		return false
	}
	kind := Classify(err)
	return kind != FailurePermanent
}
