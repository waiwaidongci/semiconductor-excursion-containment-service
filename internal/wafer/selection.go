package wafer

import (
	"errors"
	"sort"
	"time"
)

type Predicate func(Wafer) bool

func Select(wafers []Wafer, predicate Predicate) []Wafer {
	if predicate == nil {
		return CloneAll(wafers)
	}
	selected := make([]Wafer, 0, len(wafers))
	for _, item := range wafers {
		if predicate(item) {
			selected = append(selected, item.Clone())
		}
	}
	return selected
}

func SortBySlot(wafers []Wafer) []Wafer {
	sorted := CloneAll(wafers)
	sort.SliceStable(sorted, func(left, right int) bool { return sorted[left].Slot < sorted[right].Slot })
	return sorted
}

func Annotate(wafers []Wafer, key, value string) ([]Wafer, error) {
	if key == "" {
		return nil, errors.New("annotation key is required")
	}
	annotated := CloneAll(wafers)
	for index := range annotated {
		if annotated[index].Tags == nil {
			annotated[index].Tags = make(map[string]string)
		}
		annotated[index].Tags[key] = value
	}
	return annotated, nil
}

func Reclassify(wafers []Wafer, disposition string, measuredAt time.Time) ([]Wafer, error) {
	if disposition == "" {
		return nil, errors.New("disposition is required")
	}
	reclassified := CloneAll(wafers)
	for index := range reclassified {
		timestamp := measuredAt
		reclassified[index].Disposition = disposition
		reclassified[index].MeasuredAt = &timestamp
	}
	return reclassified, nil
}

func Split(wafers []Wafer, predicate Predicate) (matched, remaining []Wafer) {
	for _, item := range wafers {
		if predicate != nil && predicate(item) {
			matched = append(matched, item.Clone())
		} else {
			remaining = append(remaining, item.Clone())
		}
	}
	return matched, remaining
}

func Merge(groups ...[]Wafer) []Wafer {
	total := 0
	for _, group := range groups {
		total += len(group)
	}
	merged := make([]Wafer, 0, total)
	for _, group := range groups {
		merged = append(merged, CloneAll(group)...)
	}
	return SortBySlot(merged)
}
