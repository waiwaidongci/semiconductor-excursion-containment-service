package wafer

import (
	"reflect"
	"testing"
	"time"
)

func TestEWaferDerivationIsolation(t *testing.T) {
	now := time.Date(2026, 8, 19, 11, 0, 0, 0, time.UTC)
	source := []Wafer{
		{ID: "W2", Slot: 2, Disposition: "hold", MeasuredAt: &now, Measurements: []Measurement{{Name: "cd", Value: 42, Unit: "nm"}}, Tags: map[string]string{"origin": "fab-a"}},
		{ID: "W1", Slot: 1, Disposition: "hold", MeasuredAt: &now, Measurements: []Measurement{{Name: "cd", Value: 41, Unit: "nm"}}, Tags: map[string]string{"origin": "fab-a"}},
	}
	originalTime := now
	original := []Wafer{
		{ID: "W2", Slot: 2, Disposition: "hold", MeasuredAt: &originalTime, Measurements: []Measurement{{Name: "cd", Value: 42, Unit: "nm"}}, Tags: map[string]string{"origin": "fab-a"}},
		{ID: "W1", Slot: 1, Disposition: "hold", MeasuredAt: &originalTime, Measurements: []Measurement{{Name: "cd", Value: 41, Unit: "nm"}}, Tags: map[string]string{"origin": "fab-a"}},
	}
	selected := Select(source, func(item Wafer) bool { return item.Slot == 1 })
	selected[0].Tags["origin"] = "changed"
	selected[0].Measurements[0].Value = 999
	sorted := SortBySlot(source)
	annotated, err := Annotate(sorted, "review", "required")
	if err != nil {
		t.Fatal(err)
	}
	annotated[0].Tags["origin"] = "changed-again"
	reclassified, err := Reclassify(source, "scrap", now.Add(time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	reclassified[0].MeasuredAt = nil
	if !reflect.DeepEqual(source, original) {
		t.Fatalf("derived wafer data mutated source\nwant: %#v\n got: %#v", original, source)
	}
	if sorted[0].ID != "W1" || annotated[0].Tags["review"] != "required" || reclassified[0].Disposition != "scrap" {
		t.Fatal("derived results are incomplete")
	}
	matched, remaining := Split(original, func(item Wafer) bool { return item.Slot == 1 })
	merged := Merge(matched, remaining)
	matched[0].Tags["origin"] = "split-mutated"
	merged[0].Measurements[0].Value = -1
	if original[1].Tags["origin"] != "fab-a" || original[1].Measurements[0].Value != 41 {
		t.Fatalf("split or merge retained nested aliases: %#v", original)
	}
}
