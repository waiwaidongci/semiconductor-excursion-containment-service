package review

import (
	"errors"
	"fmt"
	"sort"
	"time"
)

var (
	ErrReviewMissing  = errors.New("engineering review missing")
	ErrReviewExpired  = errors.New("engineering review expired")
	ErrReviewConflict = errors.New("engineering review conflict")
)

type Decision string

const (
	DecisionPending Decision = "pending"
	DecisionApprove Decision = "approve"
	DecisionReject  Decision = "reject"
)

type Vote struct {
	Engineer string
	Decision Decision
	Comment  string
	VotedAt  time.Time
}

func (vote Vote) Validate() error {
	if vote.Engineer == "" || vote.Comment == "" {
		return errors.New("engineer and comment are required")
	}
	if vote.Decision != DecisionApprove && vote.Decision != DecisionReject {
		return fmt.Errorf("unsupported vote decision %q", vote.Decision)
	}
	return nil
}

type Review struct {
	ID       string
	LotID    string
	Required int
	Decision Decision
	Votes    []Vote
	OpenedAt time.Time
	Deadline time.Time
	ClosedAt *time.Time
	Revision int
}

func New(id, lotID string, required int, openedAt, deadline time.Time) (*Review, error) {
	if id == "" || lotID == "" {
		return nil, errors.New("review and lot identifiers are required")
	}
	if required < 1 {
		return nil, errors.New("at least one approval is required")
	}
	if !deadline.After(openedAt) {
		return nil, errors.New("review deadline must follow opening time")
	}
	return &Review{ID: id, LotID: lotID, Required: required, Decision: DecisionPending, Votes: make([]Vote, 0, required), OpenedAt: openedAt, Deadline: deadline, Revision: 1}, nil
}

func (review *Review) Clone() *Review {
	if review == nil {
		return nil
	}
	clone := *review
	clone.Votes = append([]Vote(nil), review.Votes...)
	if review.ClosedAt != nil {
		closedAt := *review.ClosedAt
		clone.ClosedAt = &closedAt
	}
	return &clone
}

func (review *Review) AddVote(vote Vote, now time.Time) error {
	if review == nil {
		return ErrReviewMissing
	}
	if review.Decision != DecisionPending {
		return errors.New("review is already closed")
	}
	if now.After(review.Deadline) {
		return ErrReviewExpired
	}
	if err := vote.Validate(); err != nil {
		return err
	}
	for _, existing := range review.Votes {
		if existing.Engineer == vote.Engineer {
			return fmt.Errorf("%w: engineer %s already voted", ErrReviewConflict, vote.Engineer)
		}
	}
	review.Votes = append(review.Votes, vote)
	sort.SliceStable(review.Votes, func(left, right int) bool { return review.Votes[left].Engineer < review.Votes[right].Engineer })
	review.Revision++
	return review.recalculate(now)
}

func (review *Review) recalculate(now time.Time) error {
	approvals := 0
	for _, vote := range review.Votes {
		if vote.Decision == DecisionReject {
			review.Decision = DecisionReject
			review.ClosedAt = &now
			return nil
		}
		if vote.Decision == DecisionApprove {
			approvals++
		}
	}
	if approvals >= review.Required {
		review.Decision = DecisionApprove
		review.ClosedAt = &now
	}
	return nil
}

func (review *Review) RemainingApprovals() int {
	if review == nil || review.Decision != DecisionPending {
		return 0
	}
	approvals := 0
	for _, vote := range review.Votes {
		if vote.Decision == DecisionApprove {
			approvals++
		}
	}
	remaining := review.Required - approvals
	if remaining < 0 {
		return 0
	}
	return remaining
}
