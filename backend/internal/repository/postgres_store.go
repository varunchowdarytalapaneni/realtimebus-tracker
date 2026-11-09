package repository

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"bus-tracker/internal/domain"
)

type TelemetryStore struct {
	pool *pgxpool.Pool
}

func NewTelemetryStore(pool *pgxpool.Pool) *TelemetryStore {
	return &TelemetryStore{pool: pool}
}

func (s *TelemetryStore) InsertBatch(ctx context.Context, items []domain.Telemetry) error {
	batch := &pgx.Batch{}
	for _, item := range items {
		batch.Queue(`
			INSERT INTO live_telemetry_logs (bus_id, location, speed, heading, recorded_at)
			VALUES ($1, ST_SetSRID(ST_MakePoint($2, $3), 4326), $4, $5, $6)`,
			item.BusID, item.Longitude, item.Latitude, item.Speed, item.Heading, item.Timestamp)
	}
	results := s.pool.SendBatch(ctx, batch)
	defer results.Close()
	for range items {
		if _, err := results.Exec(); err != nil {
			return err
		}
	}
	return nil
}
