package containment

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

type Repository interface {
	Get(id string) (*Lot, error)
	Save(lot *Lot, expectedRevision int) error
}

type MemoryRepository struct {
	mu   sync.RWMutex
	lots map[string]*Lot
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{lots: make(map[string]*Lot)}
}

func (repository *MemoryRepository) Get(id string) (*Lot, error) {
	repository.mu.RLock()
	defer repository.mu.RUnlock()
	lot, ok := repository.lots[id]
	if !ok {
		return nil, fmt.Errorf("load containment %s: %w", id, ErrMissingLot)
	}
	return lot.Clone(), nil
}

func (repository *MemoryRepository) Save(lot *Lot, expectedRevision int) error {
	if err := lot.Validate(); err != nil {
		return err
	}
	repository.mu.Lock()
	defer repository.mu.Unlock()
	current, exists := repository.lots[lot.ID]
	if exists && current.Revision != expectedRevision {
		return fmt.Errorf("revision changed from %d to %d", expectedRevision, current.Revision)
	}
	repository.lots[lot.ID] = lot.Clone()
	return nil
}

type Service struct {
	repository Repository
	clock      func() time.Time
}

func NewService(repository Repository, clock func() time.Time) *Service {
	if clock == nil {
		clock = time.Now
	}
	return &Service{repository: repository, clock: clock}
}

func (service *Service) Register(lot *Lot) error {
	if lot == nil {
		return ErrMissingLot
	}
	return service.repository.Save(lot.Clone(), 0)
}

func (service *Service) Advance(id string, next State, actor, reason string) (*Lot, error) {
	lot, err := service.repository.Get(id)
	if err != nil {
		return nil, err
	}
	if !CanTransition(lot.State, next) {
		return nil, fmt.Errorf("%w: %s to %s", ErrInvalidTransition, lot.State, next)
	}
	if actor == "" || reason == "" {
		return nil, errors.New("actor and reason are required")
	}
	expectedRevision := lot.Revision
	lot.History = append(lot.History, Transition{From: lot.State, To: next, Actor: actor, Reason: reason, OccurredAt: service.clock()})
	lot.State = next
	lot.Reason = reason
	lot.Revision++
	if err := service.repository.Save(lot, expectedRevision); err != nil {
		return nil, err
	}
	return lot.Clone(), nil
}

func CanTransition(current, next State) bool {
	allowed := map[State]map[State]bool{
		StateDetected:      {StateContained: true},
		StateContained:     {StateInvestigating: true, StateReleased: true},
		StateInvestigating: {StateContained: true, StateReleased: true},
		StateReleased:      {},
	}
	return allowed[current][next]
}

func (service *Service) Release(id, actor string) (*Lot, error) {
	lot, err := service.repository.Get(id)
	if err != nil {
		return nil, err
	}
	if len(lot.AffectedTools) == 0 {
		return nil, errors.New("release requires at least one assessed tool")
	}
	return service.Advance(id, StateReleased, actor, "engineering review complete")
}
