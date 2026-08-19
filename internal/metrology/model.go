package metrology

import (
	"errors"
	"sort"
	"time"
)

type Key struct {
	ToolID string
	Metric string
}

func (key Key) Validate() error {
	if key.ToolID == "" || key.Metric == "" {
		return errors.New("tool and metric are required")
	}
	return nil
}

type Point struct {
	Key        Key
	Value      float64
	Unit       string
	ObservedAt time.Time
	Sequence   uint64
}

func (point Point) Validate() error {
	if err := point.Key.Validate(); err != nil {
		return err
	}
	if point.Unit == "" {
		return errors.New("measurement unit is required")
	}
	return nil
}

type Snapshot struct {
	Points    []Point
	Version   uint64
	CreatedAt time.Time
}

func (snapshot Snapshot) Clone() Snapshot {
	return snapshot
}

func (snapshot Snapshot) Sorted() Snapshot {
	sort.SliceStable(snapshot.Points, func(left, right int) bool {
		if snapshot.Points[left].Key.ToolID == snapshot.Points[right].Key.ToolID {
			return snapshot.Points[left].Key.Metric < snapshot.Points[right].Key.Metric
		}
		return snapshot.Points[left].Key.ToolID < snapshot.Points[right].Key.ToolID
	})
	return snapshot
}

func (snapshot Snapshot) ByTool(toolID string) []Point {
	result := make([]Point, 0)
	for _, point := range snapshot.Points {
		if point.Key.ToolID == toolID {
			result = append(result, point)
		}
	}
	return result
}
