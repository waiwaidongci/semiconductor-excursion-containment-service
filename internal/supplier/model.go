package supplier

import (
	"errors"
	"fmt"
	"time"
)

var (
	ErrEvidenceMissing = errors.New("supplier evidence missing")
	ErrEvidencePending = errors.New("supplier evidence pending")
	ErrSupplierBlocked = errors.New("supplier blocked")
)

type Status string

const (
	StatusRequested Status = "requested"
	StatusPending   Status = "pending"
	StatusReady     Status = "ready"
	StatusRejected  Status = "rejected"
)

type Request struct {
	ID          string
	SupplierID  string
	LotID       string
	Status      Status
	Attempts    int
	RequestedAt time.Time
	UpdatedAt   time.Time
}

func NewRequest(id, supplierID, lotID string, now time.Time) (*Request, error) {
	if id == "" || supplierID == "" || lotID == "" {
		return nil, errors.New("request, supplier, and lot are required")
	}
	return &Request{ID: id, SupplierID: supplierID, LotID: lotID, Status: StatusRequested, RequestedAt: now, UpdatedAt: now}, nil
}

func (request *Request) Clone() *Request {
	if request == nil {
		return nil
	}
	clone := *request
	return &clone
}

func (request *Request) MarkPending(now time.Time) error {
	if request == nil {
		return errors.New("supplier evidence missing")
	}
	if request.Status == StatusRejected {
		return errors.New("supplier blocked")
	}
	request.Status = StatusPending
	request.Attempts++
	request.UpdatedAt = now
	return nil
}

func (request *Request) MarkReady(now time.Time) error {
	if request == nil {
		return ErrEvidenceMissing
	}
	if request.Status != StatusPending && request.Status != StatusRequested {
		return fmt.Errorf("cannot mark %s request ready", request.Status)
	}
	request.Status = StatusReady
	request.UpdatedAt = now
	return nil
}

type FailureKind string

const (
	FailurePermanent FailureKind = "permanent"
	FailureRetryable FailureKind = "retryable"
	FailureUnknown   FailureKind = "unknown"
)

func Classify(err error) FailureKind {
	switch {
	case err == ErrEvidenceMissing, err == ErrSupplierBlocked:
		return FailurePermanent
	case err == ErrEvidencePending:
		return FailureRetryable
	default:
		return FailureUnknown
	}
}
