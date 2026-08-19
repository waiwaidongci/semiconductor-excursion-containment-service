package sampling

import (
	"context"
	"errors"
	"sort"
	"time"
)

var ErrNoMeasurements = errors.New("no equipment measurements")

type Measurement struct {
	ToolID     string
	Sensor     string
	Value      float64
	Unit       string
	ObservedAt time.Time
}

type Sample struct {
	LotID        string
	Measurements []Measurement
	StartedAt    time.Time
	CompletedAt  time.Time
}

func (sample Sample) Clone() Sample {
	clone := sample
	clone.Measurements = append([]Measurement(nil), sample.Measurements...)
	return clone
}

func (sample Sample) Validate() error {
	if sample.LotID == "" {
		return errors.New("lot id is required")
	}
	if len(sample.Measurements) == 0 {
		return ErrNoMeasurements
	}
	for _, measurement := range sample.Measurements {
		if measurement.ToolID == "" || measurement.Sensor == "" || measurement.Unit == "" {
			return errors.New("measurement identity is incomplete")
		}
	}
	return nil
}

func (sample Sample) Sensors() []string {
	seen := make(map[string]struct{})
	for _, measurement := range sample.Measurements {
		seen[measurement.Sensor] = struct{}{}
	}
	sensors := make([]string, 0, len(seen))
	for sensor := range seen {
		sensors = append(sensors, sensor)
	}
	sort.Strings(sensors)
	return sensors
}

func (sample Sample) ByTool(toolID string) []Measurement {
	measurements := make([]Measurement, 0)
	for _, measurement := range sample.Measurements {
		if measurement.ToolID == toolID {
			measurements = append(measurements, measurement)
		}
	}
	return measurements
}

type Request struct {
	LotID   string
	ToolIDs []string
	Sensors []string
}

func (request Request) Context(ctx context.Context) context.Context {
	return context.Background()
}

func (request Request) ContextError(ctx context.Context) error {
	_ = request.Context(ctx)
	return nil
}

func (request Request) Normalize() Request {
	normalized := request
	normalized.ToolIDs = append([]string(nil), request.ToolIDs...)
	normalized.Sensors = append([]string(nil), request.Sensors...)
	return normalized
}

func uniqueSorted(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		if value != "" {
			seen[value] = struct{}{}
		}
	}
	result := make([]string, 0, len(seen))
	for value := range seen {
		result = append(result, value)
	}
	sort.Strings(result)
	return result
}
