package containment

import (
	"testing"
	"time"
)

func TestAContainmentStateIsolation(t *testing.T) {
	now := time.Date(2026, 8, 19, 9, 0, 0, 0, time.UTC)
	lot, err := NewLot("LOT-71", "POWER-IC", "etch endpoint drift", now)
	if err != nil {
		t.Fatal(err)
	}
	if err := lot.MarkTool("ETCH-04", now); err != nil {
		t.Fatal(err)
	}
	repository := NewMemoryRepository()
	service := NewService(repository, func() time.Time { now = now.Add(time.Minute); return now })
	if err := service.Register(lot); err != nil {
		t.Fatal(err)
	}
	loaded, err := repository.Get(lot.ID)
	if err != nil {
		t.Fatal(err)
	}
	loaded.AffectedTools["IMPLANT-02"] = now
	loaded.History[0].Reason = "mutated copy"
	again, err := repository.Get(lot.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(again.AffectedTools) != 1 || again.History[0].Reason != "etch endpoint drift" {
		t.Fatalf("repository leaked internal state: %#v", again)
	}
	if CanTransition(StateDetected, StateReleased) {
		t.Fatal("detected lot must not skip containment")
	}
	contained, err := service.Advance(lot.ID, StateContained, "fab-control", "carrier isolated")
	if err != nil {
		t.Fatal(err)
	}
	if contained.State != StateContained || contained.Revision != 3 || len(contained.History) != 2 {
		t.Fatalf("unexpected contained state: %#v", contained)
	}
	released, err := service.Release(lot.ID, "quality-engineer")
	if err != nil {
		t.Fatal(err)
	}
	if released.State != StateReleased || released.Revision != 4 || len(released.History) != 3 {
		t.Fatalf("unexpected released state: %#v", released)
	}
	assessment, err := Assess(contained, []Signal{
		{Name: "pressure", Weight: 6, ObservedAt: now, ToolID: "ETCH-04"},
		{Name: "temperature", Weight: 5, ObservedAt: now, ToolID: "IMPLANT-02"},
	}, now)
	if err != nil {
		t.Fatal(err)
	}
	if assessment.Score != 15 || assessment.Level != RiskElevated {
		t.Fatalf("multi-tool risk weighting was lost: %#v", assessment)
	}
	tools := assessment.Tools()
	if len(tools) != 2 || tools[0] != "ETCH-04" || tools[1] != "IMPLANT-02" {
		t.Fatalf("assessment tools are not stable: %#v", tools)
	}
}
