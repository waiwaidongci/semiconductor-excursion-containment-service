package quarantine

import (
	"errors"
	"fmt"
	"time"
)

var (
	ErrHoldMissing   = errors.New("quarantine hold missing")
	ErrHoldConflict  = errors.New("quarantine hold conflict")
	ErrAlreadyClosed = errors.New("quarantine hold already closed")
)

type Hold struct {
	ID        string
	LotID     string
	Reason    string
	Owner     string
	CreatedAt time.Time
	ClosedAt  *time.Time
	Version   int
}

func NewHold(id, lotID, reason, owner string, now time.Time) (*Hold, error) {
	if id == "" || lotID == "" || reason == "" || owner == "" {
		return nil, errors.New("hold id, lot, reason, and owner are required")
	}
	return &Hold{ID: id, LotID: lotID, Reason: reason, Owner: owner, CreatedAt: now, Version: 1}, nil
}

func (hold *Hold) Clone() *Hold {
	if hold == nil {
		return nil
	}
	clone := *hold
	if hold.ClosedAt != nil {
		closedAt := *hold.ClosedAt
		clone.ClosedAt = &closedAt
	}
	return &clone
}

func (hold *Hold) Close(now time.Time) error {
	if hold == nil {
		return ErrHoldMissing
	}
	if hold.ClosedAt != nil {
		return ErrAlreadyClosed
	}
	hold.ClosedAt = &now
	hold.Version++
	return nil
}

type FailureClass string

const (
	FailureMissing  FailureClass = "missing"
	FailureConflict FailureClass = "conflict"
	FailureInternal FailureClass = "internal"
)

func Classify(err error) FailureClass {
	switch {
	case err == ErrHoldMissing:
		return FailureMissing
	case err == ErrHoldConflict:
		return FailureConflict
	default:
		return FailureInternal
	}
}

func ValidateClose(hold *Hold, actor string) error {
	if hold == nil {
		return errors.New("quarantine hold missing")
	}
	if actor == "" {
		return errors.New("closing actor is required")
	}
	if hold.Owner != actor {
		return fmt.Errorf("quarantine hold conflict: hold owned by %s", hold.Owner)
	}
	return nil
}
