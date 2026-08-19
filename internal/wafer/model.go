package wafer

import (
	"errors"
	"fmt"
	"sort"
	"time"
)

type Measurement struct {
	Name  string
	Value float64
	Unit  string
}

type Wafer struct {
	ID           string
	Slot         int
	Disposition  string
	MeasuredAt   *time.Time
	Measurements []Measurement
	Tags         map[string]string
}

func (wafer Wafer) Clone() Wafer {
	clone := wafer
	if wafer.MeasuredAt != nil {
		measuredAt := *wafer.MeasuredAt
		clone.MeasuredAt = &measuredAt
	}
	clone.Measurements = append([]Measurement(nil), wafer.Measurements...)
	clone.Tags = make(map[string]string, len(wafer.Tags))
	for key, value := range wafer.Tags {
		clone.Tags[key] = value
	}
	return clone
}

func (wafer Wafer) Validate() error {
	if wafer.ID == "" {
		return errors.New("wafer id is required")
	}
	if wafer.Slot < 1 || wafer.Slot > 25 {
		return fmt.Errorf("wafer slot %d is outside carrier range", wafer.Slot)
	}
	for _, measurement := range wafer.Measurements {
		if measurement.Name == "" || measurement.Unit == "" {
			return errors.New("measurement name and unit are required")
		}
	}
	return nil
}

type Snapshot struct {
	LotID     string
	Wafers    []Wafer
	CreatedAt time.Time
	Reason    string
}

func NewSnapshot(lotID string, wafers []Wafer, reason string, now time.Time) (Snapshot, error) {
	if lotID == "" || reason == "" {
		return Snapshot{}, errors.New("lot id and snapshot reason are required")
	}
	cloned := CloneAll(wafers)
	for _, item := range cloned {
		if err := item.Validate(); err != nil {
			return Snapshot{}, err
		}
	}
	sort.SliceStable(cloned, func(left, right int) bool { return cloned[left].Slot < cloned[right].Slot })
	return Snapshot{LotID: lotID, Wafers: cloned, CreatedAt: now, Reason: reason}, nil
}

func (snapshot Snapshot) Clone() Snapshot {
	clone := snapshot
	clone.Wafers = CloneAll(snapshot.Wafers)
	return clone
}

func CloneAll(wafers []Wafer) []Wafer {
	cloned := make([]Wafer, len(wafers))
	for index := range wafers {
		cloned[index] = wafers[index].Clone()
	}
	return cloned
}
