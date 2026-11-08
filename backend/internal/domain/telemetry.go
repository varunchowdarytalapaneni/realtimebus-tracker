package domain

import (
	"errors"
	"time"
)

var ErrInvalidTelemetry = errors.New("invalid telemetry")

type Telemetry struct {
	BusID     int64     `json:"bus_id"`
	Latitude  float64   `json:"latitude"`
	Longitude float64   `json:"longitude"`
	Speed     float64   `json:"speed"`
	Heading   float64   `json:"heading"`
	Timestamp time.Time `json:"timestamp"`
}

func (t Telemetry) Validate(now time.Time) error {
	if t.BusID <= 0 || t.Latitude < -90 || t.Latitude > 90 || t.Longitude < -180 || t.Longitude > 180 {
		return ErrInvalidTelemetry
	}
	if t.Timestamp.IsZero() || t.Timestamp.After(now.Add(time.Minute)) {
		return ErrInvalidTelemetry
	}
	return nil
}
