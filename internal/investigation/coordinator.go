package investigation

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

type Inspector interface {
	Inspect(ctx context.Context, lotID string, chamber Chamber) (Finding, error)
}

type Coordinator struct {
	inspector Inspector
	now       func() time.Time
}

func NewCoordinator(inspector Inspector, now func() time.Time) *Coordinator {
	if now == nil {
		now = time.Now
	}
	return &Coordinator{inspector: inspector, now: now}
}

func (coordinator *Coordinator) Run(ctx context.Context, lotID string, chambers []Chamber, parallelism int) (Report, error) {
	if lotID == "" || len(chambers) == 0 {
		return Report{}, errors.New("lot and chambers are required")
	}
	for _, chamber := range chambers {
		if err := chamber.Validate(); err != nil {
			return Report{}, err
		}
	}
	if parallelism < 1 {
		parallelism = 1
	}
	startedAt := coordinator.now()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	type result struct {
		finding Finding
		err     error
	}
	results := make(chan result, len(chambers))
	semaphore := make(chan struct{}, parallelism)
	var waitGroup sync.WaitGroup
	for _, chamber := range chambers {
		chamber := chamber
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			select {
			case semaphore <- struct{}{}:
			case <-ctx.Done():
				results <- result{err: ctx.Err()}
				return
			}
			defer func() { <-semaphore }()
			finding, err := coordinator.inspector.Inspect(ctx, lotID, chamber)
			results <- result{finding: finding, err: err}
		}()
	}
	go func() {
		waitGroup.Wait()
		close(results)
	}()
	findings := make([]Finding, 0, len(chambers))
	for item := range results {
		if item.err != nil {
			cancel()
			return Report{}, fmt.Errorf("inspect chamber: %w", item.err)
		}
		if err := item.finding.Validate(); err != nil {
			cancel()
			return Report{}, err
		}
		findings = append(findings, item.finding)
	}
	return Report{LotID: lotID, Findings: findings, StartedAt: startedAt, FinishedAt: coordinator.now()}.Sorted(), nil
}

func Summarize(report Report) map[string]int {
	summary := make(map[string]int)
	for _, finding := range report.Findings {
		summary[finding.Severity]++
	}
	return summary
}
