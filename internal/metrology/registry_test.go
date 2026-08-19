package metrology

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestGMetrologySnapshotLocking(t *testing.T) {
	registry := NewRegistry(time.Now)
	for index := 0; index < 32; index++ {
		point := Point{Key: Key{ToolID: "ETCH-1", Metric: fmt.Sprintf("m-%02d", index)}, Value: float64(index), Unit: "nm", ObservedAt: time.Now()}
		if err := registry.Put(point); err != nil {
			t.Fatal(err)
		}
	}
	start := make(chan struct{})
	var waitGroup sync.WaitGroup
	for worker := 0; worker < 4; worker++ {
		waitGroup.Add(1)
		go func(worker int) {
			defer waitGroup.Done()
			<-start
			for iteration := 0; iteration < 200; iteration++ {
				key := Key{ToolID: "ETCH-1", Metric: fmt.Sprintf("m-%02d", iteration%32)}
				_ = registry.Put(Point{Key: key, Value: float64(worker*1000 + iteration), Unit: "nm", ObservedAt: time.Now()})
			}
		}(worker)
	}
	waitGroup.Add(1)
	go func() {
		defer waitGroup.Done()
		<-start
		for iteration := 0; iteration < 100; iteration++ {
			points := make([]Point, 0, 32)
			for index := 0; index < 32; index++ {
				points = append(points, Point{Key: Key{ToolID: "ETCH-1", Metric: fmt.Sprintf("m-%02d", index)}, Value: float64(iteration), Unit: "nm", ObservedAt: time.Now()})
			}
			if err := registry.Replace(points); err != nil {
				t.Errorf("replace failed: %v", err)
				return
			}
		}
	}()
	waitGroup.Add(1)
	go func() {
		defer waitGroup.Done()
		<-start
		for iteration := 0; iteration < 200; iteration++ {
			snapshot := registry.Snapshot()
			if len(snapshot.Points) != 32 {
				t.Errorf("snapshot lost points: %d", len(snapshot.Points))
				return
			}
			_ = registry.Count()
			_, _ = registry.Get(Key{ToolID: "ETCH-1", Metric: "m-00"})
		}
	}()
	close(start)
	waitGroup.Wait()
	snapshot := registry.Snapshot()
	snapshot.Points[0].Value = -999
	again := registry.Snapshot()
	if again.Points[0].Value == -999 || registry.Count() != 32 {
		t.Fatal("snapshot exposed registry state")
	}
	pointSource := []Point{{Key: Key{ToolID: "B", Metric: "z"}, Unit: "nm"}, {Key: Key{ToolID: "A", Metric: "a"}, Unit: "nm"}}
	modelSnapshot := Snapshot{Points: pointSource}
	cloned := modelSnapshot.Clone()
	sorted := modelSnapshot.Sorted()
	cloned.Points[0].Value = 55
	sorted.Points[0].Value = 77
	if pointSource[0].Value != 0 || pointSource[1].Value != 0 {
		t.Fatal("snapshot clone or sorting mutated the model source")
	}
}
