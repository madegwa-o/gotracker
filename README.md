# Edge Telemetry Gateway (Go)

Lightweight WebSocket telemetry gateway for ARM edge devices (e.g., Raspberry Pi 4).

## Features

- `/ws/device` WebSocket endpoint for vehicle telemetry ingest.
- Outbound cloud forwarding via WebSocket client.
- Telemetry validation (id/lat/lng/s/t).
- Bounded in-memory FIFO buffer when cloud is unavailable.
- Automatic cloud reconnect loop.
- Structured logging with zerolog.
- Environment-driven config with viper.

## Run

```bash
go mod tidy
go run ./cmd/server
```

## Configuration

Environment variables are prefixed with `EDGE_`.

- `EDGE_HTTP_ADDR` (default `:8080`)
- `EDGE_CLOUD_WS_URL` (default `ws://localhost:9090/ws/cloud`)
- `EDGE_BUFFER_SIZE` (default `10000`)
- `EDGE_RECONNECT_DELAY` (default `3s`)
- `EDGE_FLUSH_INTERVAL` (default `1s`)
- `EDGE_WRITE_TIMEOUT` (default `5s`)
- `EDGE_LOG_LEVEL` (`debug|info|warn|error`, default `info`)

## Sample telemetry packet

```json
{
  "id": "17",
  "lat": -1.2921,
  "lng": 36.8219,
  "s": 42,
  "t": 1710240012
}
```

Send this payload once per second from each vehicle client to `ws://<gateway-host>:8080/ws/device`.
