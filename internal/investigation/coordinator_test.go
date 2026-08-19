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

type errorInspector struct{}

func (errorInspector) Inspect(ctx context.Context, lotID string, chamber Chamber) (Finding, error) {
	return Finding{}, fmt.Errorf("inspection failed for %s", chamber.ChamberID)
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

func TestIInvestigationFanoutLifecycle(t *testing.T) {
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
	start := make(chan struct{})
	done := make(chan outcome, 2)
	for run := 0; run < 2; run++ {
		go func() {
			<-start
			report, err := coordinator.Run(context.Background(), "LOT-I", chambers, len(chambers))
			done <- outcome{report: report, err: err}
		}()
	}
	close(start)
	deadline := time.After(2 * time.Second)
	for inspector.started.Load() != int32(len(chambers)*2) {
		select {
		case <-deadline:
			t.Fatal("not all investigation workers started")
		default:
			time.Sleep(time.Millisecond)
		}
	}
	close(release)
	for run := 0; run < 2; run++ {
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
	modelSource := []Finding{{Chamber: Chamber{ToolID: "B", ChamberID: "2", RecipeID: "R"}, Severity: "warning", Summary: "b"}, {Chamber: Chamber{ToolID: "A", ChamberID: "1", RecipeID: "R"}, Severity: "warning", Summary: "a"}}
	modelReport := Report{LotID: "LOT-I", Findings: modelSource}
	cloned := modelReport.Clone()
	sorted := modelReport.Sorted()
	cloned.Findings[0].Summary = "changed"
	sorted.Findings[0].Summary = "changed-sorted"
	if modelSource[0].Summary != "b" || modelSource[1].Summary != "a" {
		t.Fatal("report clone or sorting mutated source findings")
	}
	if _, err := NewCoordinator(errorInspector{}, time.Now).Run(context.Background(), "LOT-I", chambers[:2], 2); err == nil {
		t.Fatal("coordinator swallowed chamber inspection errors")
	}
}
