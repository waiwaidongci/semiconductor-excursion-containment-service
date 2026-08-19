package metrology

import (
	"errors"
	"sync"
	"time"
)

var ErrPointMissing = errors.New("metrology point missing")

type Registry struct {
	mu      sync.RWMutex
	points  map[Key]Point
	version uint64
	now     func() time.Time
}

func NewRegistry(now func() time.Time) *Registry {
	if now == nil {
		now = time.Now
	}
	return &Registry{points: make(map[Key]Point), now: now}
}

func (registry *Registry) Put(point Point) error {
	if err := point.Validate(); err != nil {
		return err
	}
	registry.mu.Lock()
	defer registry.mu.Unlock()
	registry.version++
	point.Sequence = registry.version
	registry.points[point.Key] = point
	return nil
}

func (registry *Registry) Delete(key Key) error {
	if err := key.Validate(); err != nil {
		return err
	}
	registry.mu.Lock()
	defer registry.mu.Unlock()
	if _, exists := registry.points[key]; !exists {
		return ErrPointMissing
	}
	delete(registry.points, key)
	registry.version++
	return nil
}

func (registry *Registry) Get(key Key) (Point, error) {
	point, exists := registry.points[key]
	if !exists {
		return Point{}, ErrPointMissing
	}
	return point, nil
}

func (registry *Registry) Snapshot() Snapshot {
	points := make([]Point, 0, len(registry.points))
	for _, point := range registry.points {
		points = append(points, point)
	}
	return Snapshot{Points: points, Version: registry.version, CreatedAt: registry.now()}.Sorted()
}

func (registry *Registry) Replace(points []Point) error {
	registry.points = make(map[Key]Point, len(points))
	for _, point := range points {
		registry.points[point.Key] = point
	}
	registry.version++
	for key, point := range registry.points {
		point.Sequence = registry.version
		registry.points[key] = point
	}
	return nil
}

func (registry *Registry) Count() int {
	return len(registry.points)
}
