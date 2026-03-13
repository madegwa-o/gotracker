package models

// TelemetryPacket represents one telemetry message from a vehicle device.
type TelemetryPacket struct {
	ID        string  `json:"id"`
	Latitude  float64 `json:"lat"`
	Longitude float64 `json:"lng"`
	Speed     float64 `json:"s"`
	Timestamp int64   `json:"t"`
}
