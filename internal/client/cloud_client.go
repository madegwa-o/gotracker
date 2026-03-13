package client

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/rs/zerolog"
)

// CloudClient manages outbound websocket connection to cloud ingestion.
type CloudClient struct {
	url            string
	reconnectDelay time.Duration
	writeTimeout   time.Duration
	log            zerolog.Logger

	mu   sync.RWMutex
	conn *websocket.Conn
}

func NewCloudClient(url string, reconnectDelay, writeTimeout time.Duration, log zerolog.Logger) *CloudClient {
	if reconnectDelay <= 0 {
		reconnectDelay = 3 * time.Second
	}
	if writeTimeout <= 0 {
		writeTimeout = 5 * time.Second
	}
	return &CloudClient{url: url, reconnectDelay: reconnectDelay, writeTimeout: writeTimeout, log: log}
}

func (c *CloudClient) Start(ctx context.Context) {
	go c.connectLoop(ctx)
}

func (c *CloudClient) connectLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			c.closeConn()
			return
		default:
		}

		if c.isConnected() {
			time.Sleep(250 * time.Millisecond)
			continue
		}

		conn, _, err := websocket.DefaultDialer.DialContext(ctx, c.url, nil)
		if err != nil {
			c.log.Error().Err(err).Str("url", c.url).Msg("cloud websocket dial failed")
			select {
			case <-ctx.Done():
				return
			case <-time.After(c.reconnectDelay):
			}
			continue
		}

		c.mu.Lock()
		c.conn = conn
		c.mu.Unlock()
		c.log.Info().Str("url", c.url).Msg("connected to cloud websocket")

		if err := c.readUntilClose(ctx, conn); err != nil {
			c.log.Warn().Err(err).Msg("cloud websocket disconnected")
		}
		c.closeConn()
	}
}

func (c *CloudClient) readUntilClose(ctx context.Context, conn *websocket.Conn) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		if _, _, err := conn.ReadMessage(); err != nil {
			return err
		}
	}
}

func (c *CloudClient) Send(msg []byte) error {
	conn := c.getConn()
	if conn == nil {
		return errors.New("cloud websocket not connected")
	}
	if err := conn.SetWriteDeadline(time.Now().Add(c.writeTimeout)); err != nil {
		return err
	}
	if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
		c.closeConn()
		return err
	}
	return nil
}

func (c *CloudClient) IsConnected() bool { return c.isConnected() }

func (c *CloudClient) isConnected() bool {
	return c.getConn() != nil
}

func (c *CloudClient) getConn() *websocket.Conn {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.conn
}

func (c *CloudClient) closeConn() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.conn != nil {
		_ = c.conn.Close()
		c.conn = nil
	}
}
