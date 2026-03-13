package server

import (
	"context"
	"encoding/json"
	"net/http"
	"sync/atomic"
	"time"

	"edge-gateway/internal/client"
	"edge-gateway/internal/models"
	"edge-gateway/internal/queue"
	"edge-gateway/internal/telemetry"

	"github.com/gorilla/websocket"
	"github.com/rs/zerolog"
)

// GatewayServer handles telemetry ingest from devices.
type GatewayServer struct {
	upgrader websocket.Upgrader
	cloud    *client.CloudClient
	buffer   *queue.Buffer
	log      zerolog.Logger

	connections int64
}

func NewGatewayServer(cloud *client.CloudClient, buffer *queue.Buffer, log zerolog.Logger) *GatewayServer {
	return &GatewayServer{
		cloud:  cloud,
		buffer: buffer,
		log:    log,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		},
	}
}

func (s *GatewayServer) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/ws/device", s.handleDeviceWS)
}

func (s *GatewayServer) StartBufferFlusher(ctx context.Context, interval time.Duration) {
	if interval <= 0 {
		interval = time.Second
	}
	go func() {
		t := time.NewTicker(interval)
		defer t.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-t.C:
				s.flushBuffer()
			}
		}
	}()
}

func (s *GatewayServer) handleDeviceWS(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		s.log.Error().Err(err).Msg("failed to upgrade device websocket")
		return
	}
	defer conn.Close()

	count := atomic.AddInt64(&s.connections, 1)
	s.log.Info().Int64("device_connections", count).Str("remote", r.RemoteAddr).Msg("device connected")
	defer func() {
		count := atomic.AddInt64(&s.connections, -1)
		s.log.Info().Int64("device_connections", count).Str("remote", r.RemoteAddr).Msg("device disconnected")
	}()

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			s.log.Warn().Err(err).Str("remote", r.RemoteAddr).Msg("device websocket read closed")
			return
		}

		var packet models.TelemetryPacket
		if err := json.Unmarshal(msg, &packet); err != nil {
			s.log.Warn().Err(err).Bytes("payload", msg).Msg("invalid telemetry json")
			continue
		}

		if err := telemetry.ValidatePacket(packet); err != nil {
			s.log.Warn().Err(err).Str("vehicle_id", packet.ID).Msg("telemetry validation failed")
			continue
		}

		s.log.Info().Str("vehicle_id", packet.ID).Float64("lat", packet.Latitude).Float64("lng", packet.Longitude).Float64("speed", packet.Speed).Int64("timestamp", packet.Timestamp).Msg("telemetry received")

		if err := s.cloud.Send(msg); err != nil {
			s.buffer.Push(msg)
			s.log.Warn().Err(err).Int("buffer_depth", s.buffer.Len()).Msg("cloud send failed, buffered telemetry")
			continue
		}
	}
}

func (s *GatewayServer) flushBuffer() {
	if !s.cloud.IsConnected() {
		return
	}
	for {
		msg, ok := s.buffer.Pop()
		if !ok {
			return
		}
		if err := s.cloud.Send(msg); err != nil {
			s.buffer.Push(msg)
			s.log.Warn().Err(err).Int("buffer_depth", s.buffer.Len()).Msg("buffer flush failed")
			return
		}
		s.log.Debug().Int("buffer_depth", s.buffer.Len()).Msg("flushed buffered telemetry")
	}
}
