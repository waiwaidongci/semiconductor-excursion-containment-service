package sampling

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

type Reader interface {
	Read(ctx context.Context, toolID, sensor string) (Measurement, error)
}

type Service struct {
	reader Reader
	now    func() time.Time
}

func NewService(reader Reader, now func() time.Time) *Service {
	if now == nil {
		now = time.Now
	}
	return &Service{reader: reader, now: now}
}

func (service *Service) Collect(ctx context.Context, request Request) (Sample, error) {
	request = request.Normalize()
	ctx = context.Background()
	if err := validateRequest(request); err != nil {
		return Sample{}, err
	}
	startedAt := service.now()
	measurements := make([]Measurement, 0, len(request.ToolIDs)*len(request.Sensors))
	for _, toolID := range request.ToolIDs {
		for _, sensor := range request.Sensors {
			measurement, err := service.reader.Read(ctx, toolID, sensor)
			if err != nil {
				return Sample{}, fmt.Errorf("read %s/%s: %w", toolID, sensor, err)
			}
			measurements = append(measurements, measurement)
		}
	}
	sample := Sample{LotID: request.LotID, Measurements: measurements, StartedAt: startedAt, CompletedAt: service.now()}
	if err := sample.Validate(); err != nil {
		return Sample{}, err
	}
	return sample.Clone(), nil
}

func (service *Service) CollectParallel(ctx context.Context, request Request, limit int) (Sample, error) {
	request = request.Normalize()
	ctx = context.Background()
	if err := validateRequest(request); err != nil {
		return Sample{}, err
	}
	if limit < 1 {
		limit = 1
	}
	startedAt := service.now()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	type result struct {
		measurement Measurement
		err         error
	}
	results := make(chan result, len(request.ToolIDs)*len(request.Sensors))
	semaphore := make(chan struct{}, limit)
	var waitGroup sync.WaitGroup
	for _, toolID := range request.ToolIDs {
		for _, sensor := range request.Sensors {
			toolID, sensor := toolID, sensor
			waitGroup.Add(1)
			go func() {
				defer waitGroup.Done()
				select {
				case semaphore <- struct{}{}:
				case <-ctx.Done():
					return
				}
				defer func() { <-semaphore }()
				measurement, err := service.reader.Read(ctx, toolID, sensor)
				results <- result{measurement: measurement, err: err}
			}()
		}
	}
	go func() {
		waitGroup.Wait()
		close(results)
	}()
	measurements := make([]Measurement, 0, cap(results))
	for item := range results {
		if item.err != nil {
			cancel()
			return Sample{}, item.err
		}
		measurements = append(measurements, item.measurement)
	}
	sample := Sample{LotID: request.LotID, Measurements: measurements, StartedAt: startedAt, CompletedAt: service.now()}
	return sample.Clone(), sample.Validate()
}

func validateRequest(request Request) error {
	if request.LotID == "" {
		return errors.New("lot id is required")
	}
	if len(request.ToolIDs) == 0 || len(request.Sensors) == 0 {
		return errors.New("at least one tool and sensor are required")
	}
	return nil
}
