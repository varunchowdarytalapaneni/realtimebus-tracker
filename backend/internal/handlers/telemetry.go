package handlers

import (
	"context"
	"fmt"
	"log/slog"
	"sort"
	"time"

	"bus-tracker/internal/domain"
)

// Interfaces are defined here (consumer side) so the service never imports Redis/Postgres.
type Cache interface {
	SetLatestIfNewer(ctx context.Context, t domain.Telemetry) (bool, error)
	GetLatest(ctx context.Context, busID int64) (*domain.Telemetry, error)
	Publish(ctx context.Context, t domain.Telemetry) error
}

type Store interface {
	InsertBatch(ctx context.Context, items []domain.Telemetry) error
}

const (
	persistQueueSize = 10000
	persistBatchSize = 200
	persistInterval  = time.Second
)

type TelemetryService struct {
	cache     Cache
	store     Store
	persistCh chan domain.Telemetry
}

func NewTelemetryService(cache Cache, store Store) *TelemetryService {
	return &TelemetryService{
		cache:     cache,
		store:     store,
		persistCh: make(chan domain.Telemetry, persistQueueSize),
	}
}

// Ingest validates pings, updates the live view (Redis), and queues them for history (Postgres).
func (s *TelemetryService) Ingest(ctx context.Context, points []domain.Telemetry) error {
	now := time.Now()
	for i := range points {
		if err := points[i].Validate(now); err != nil {
			return fmt.Errorf("point %d: %w", i, err)
		}
	}

	sort.Slice(points, func(i, j int) bool { return points[i].Timestamp.Before(points[j].Timestamp) })

	// Only the newest ping per bus goes to the live path, so a flushed offline
	// batch doesn't "replay" the whole trip to viewers.
	newest := make(map[int64]domain.Telemetry)
	for _, p := range points {
		newest[p.BusID] = p
	}
	for _, p := range newest {
		updated, err := s.cache.SetLatestIfNewer(ctx, p)
		if err != nil {
			return fmt.Errorf("cache: %w", err)
		}
		if updated {
			if err := s.cache.Publish(ctx, p); err != nil {
				slog.Warn("publish failed", "bus_id", p.BusID, "err", err)
			}
		}
	}

	// Every ping goes to history, without blocking the request.
	for _, p := range points {
		select {
		case s.persistCh <- p:
		default:
			slog.Warn("persist queue full, dropping ping", "bus_id", p.BusID)
		}
	}
	return nil
}

func (s *TelemetryService) Latest(ctx context.Context, busID int64) (*domain.Telemetry, error) {
	return s.cache.GetLatest(ctx, busID)
}

// RunPersister drains the queue and writes to Postgres in batches. Run as a goroutine.
func (s *TelemetryService) RunPersister(ctx context.Context) {
	ticker := time.NewTicker(persistInterval)
	defer ticker.Stop()
	buf := make([]domain.Telemetry, 0, persistBatchSize)

	flush := func() {
		if len(buf) == 0 {
			return
		}
		c, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := s.store.InsertBatch(c, buf); err != nil {
			slog.Error("persist batch failed", "n", len(buf), "err", err)
		}
		buf = buf[:0]
	}

	for {
		select {
		case <-ctx.Done():
			flush()
			return
		case p := <-s.persistCh:
			buf = append(buf, p)
			if len(buf) >= persistBatchSize {
				flush()
			}
		case <-ticker.C:
			flush()
		}
	}
}
