package investigation

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"
	"time"
)

type barrierInspector struct {
	started atomic.Int32
	release <-chan struct{}
}

func (inspector *barrierInspector) Inspect(ctx context.Context, lotID string, chamber Chamber) (Finding, error) {
	inspector.started.Add(1)
	select {
	case <-inspector.release:
		return Finding{Chamber: chamber, Severity: "warning", Summary: "pressure drift", ObservedAt: time.Now()}, nil
	case <-ctx.Done():
		return Finding{}, ctx.Err()
	}
}

func TestCoordinatorWaitsForEveryInvestigationResult(t *testing.T) {
	release := make(chan struct{})
	inspector := &barrierInspector{release: release}
	coordinator := NewCoordinator(inspector, time.Now)
	chambers := make([]Chamber, 0, 8)
	for index := 0; index < 8; index++ {
		chambers = append(chambers, Chamber{ToolID: "ETCH-9", ChamberID: fmt.Sprintf("C-%d", index), RecipeID: "RCP-2"})
	}
	type outcome struct {
		report Report
		err    error
	}
	done := make(chan outcome, 1)
	go func() {
		report, err := coordinator.Run(context.Background(), "LOT-I", chambers, len(chambers))
		done <- outcome{report: report, err: err}
	}()
	deadline := time.After(2 * time.Second)
	for inspector.started.Load() != int32(len(chambers)) {
		select {
		case <-deadline:
			t.Fatal("not all investigation workers started")
		default:
			time.Sleep(time.Millisecond)
		}
	}
	close(release)
	select {
	case result := <-done:
		if result.err != nil {
			t.Fatal(result.err)
		}
		if len(result.report.Findings) != len(chambers) || Summarize(result.report)["warning"] != len(chambers) {
			t.Fatalf("coordinator returned partial report: %#v", result.report)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("coordinator did not close its result lifecycle")
	}
}
