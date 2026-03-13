package telemetry

import (
	"errors"
	"strings"
	"time"

	"edge-gateway/internal/models"
)

var (
	ErrMissingVehicleID = errors.New("vehicle id is required")
	ErrInvalidLatitude  = errors.New("latitude must be between -90 and 90")
	ErrInvalidLongitude = errors.New("longitude must be between -180 and 180")
	ErrInvalidSpeed     = errors.New("speed cannot be negative")
	ErrInvalidTimestamp = errors.New("timestamp must be a valid unix timestamp")
)

// ValidatePacket performs basic range and sanity checks.
func ValidatePacket(p models.TelemetryPacket) error {
	if strings.TrimSpace(p.ID) == "" {
		return ErrMissingVehicleID
	}
	if p.Latitude < -90 || p.Latitude > 90 {
		return ErrInvalidLatitude
	}
	if p.Longitude < -180 || p.Longitude > 180 {
		return ErrInvalidLongitude
	}
	if p.Speed < 0 {
		return ErrInvalidSpeed
	}

	now := time.Now().Unix()
	if p.Timestamp <= 0 || p.Timestamp > now+300 || p.Timestamp < now-31536000 {
		return ErrInvalidTimestamp
	}
	return nil
}
