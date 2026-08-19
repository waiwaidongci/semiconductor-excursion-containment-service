package containment

import (
	"errors"
	"fmt"
	"time"
)

var (
	ErrInvalidTransition = errors.New("invalid containment transition")
	ErrMissingLot        = errors.New("wafer lot is missing")
)

type State string

const (
	StateDetected      State = "detected"
	StateContained     State = "contained"
	StateInvestigating State = "investigating"
	StateReleased      State = "released"
)

type Lot struct {
	ID            string
	Product       string
	State         State
	Revision      int
	Reason        string
	AffectedTools map[string]time.Time
	History       []Transition
}

type Transition struct {
	From       State
	To         State
	Actor      string
	Reason     string
	OccurredAt time.Time
}

func NewLot(id, product, reason string, now time.Time) (*Lot, error) {
	if id == "" || product == "" {
		return nil, fmt.Errorf("%w: id and product are required", ErrMissingLot)
	}
	lot := &Lot{
		ID:            id,
		Product:       product,
		State:         StateDetected,
		Revision:      1,
		Reason:        reason,
		AffectedTools: make(map[string]time.Time),
		History:       make([]Transition, 0, 4),
	}
	lot.History = append(lot.History, Transition{To: StateDetected, Actor: "detector", Reason: reason, OccurredAt: now})
	return lot, nil
}

func (lot *Lot) Clone() *Lot {
	if lot == nil {
		return nil
	}
	clone := *lot
	clone.AffectedTools = make(map[string]time.Time, len(lot.AffectedTools))
	for tool, seenAt := range lot.AffectedTools {
		clone.AffectedTools[tool] = seenAt
	}
	clone.History = append([]Transition(nil), lot.History...)
	return &clone
}

func (lot *Lot) MarkTool(toolID string, observedAt time.Time) error {
	if lot == nil {
		return ErrMissingLot
	}
	if toolID == "" {
		return errors.New("tool id is required")
	}
	if lot.AffectedTools == nil {
		lot.AffectedTools = make(map[string]time.Time)
	}
	lot.AffectedTools[toolID] = observedAt
	lot.Revision++
	return nil
}

func (lot *Lot) Validate() error {
	if lot == nil || lot.ID == "" {
		return ErrMissingLot
	}
	if lot.Revision < 1 {
		return errors.New("revision must be positive")
	}
	if !knownState(lot.State) {
		return fmt.Errorf("unknown containment state %q", lot.State)
	}
	return nil
}

func knownState(state State) bool {
	switch state {
	case StateDetected, StateContained, StateInvestigating, StateReleased:
		return true
	default:
		return false
	}
}
