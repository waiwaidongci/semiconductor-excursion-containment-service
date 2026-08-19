package sampling

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

type blockingReader struct{ calls atomic.Int32 }

func (reader *blockingReader) Read(ctx context.Context, toolID, sensor string) (Measurement, error) {
	reader.calls.Add(1)
	<-ctx.Done()
	return Measurement{}, ctx.Err()
}

func TestCSamplingCancelPropagation(t *testing.T) {
	request := Request{LotID: "LOT-CTX", ToolIDs: []string{"ETCH-1", "ETCH-2"}, Sensors: []string{"pressure"}}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if !errors.Is(request.Context(ctx).Err(), context.Canceled) {
		t.Fatal("request context detached cancellation")
	}
	if !errors.Is(request.ContextError(ctx), context.Canceled) {
		t.Fatal("request context error was hidden")
	}
	reader := &blockingReader{}
	service := NewService(reader, time.Now)
	if _, err := service.Collect(ctx, request); !errors.Is(err, context.Canceled) {
		t.Fatalf("serial collection ignored cancellation: %v", err)
	}
	if _, err := service.CollectParallel(ctx, request, 2); !errors.Is(err, context.Canceled) {
		t.Fatalf("parallel collection ignored cancellation: %v", err)
	}
	if reader.calls.Load() != 0 {
		t.Fatalf("reader started %d calls after cancellation", reader.calls.Load())
	}
	normalized := (Request{ToolIDs: []string{"B", "A", "B"}, Sensors: []string{"z", "a", "z"}}).Normalize()
	if len(normalized.ToolIDs) != 2 || normalized.ToolIDs[0] != "A" || normalized.ToolIDs[1] != "B" || len(normalized.Sensors) != 2 || normalized.Sensors[0] != "a" {
		t.Fatalf("sampling request normalization is unstable: %#v", normalized)
	}
}
