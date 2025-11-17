package ws

import (
	"log/slog"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type Client struct {
	ID           string
	UUID         string
	Conn         *websocket.Conn
	Send         chan []byte
	UserAgent    string
	LastIP       string
	hub          *Hub
	done         chan struct{}
	mu           sync.RWMutex
	isActive     bool
	lastActivity time.Time
}

func NewClient(
	id string,
	uuid string,
	conn *websocket.Conn,
	userAgent string,
	lastIp string,
	hub *Hub,
) *Client {
	return &Client{
		ID:           id,
		UUID:         uuid,
		Conn:         conn,
		Send:         make(chan []byte, 256),
		UserAgent:    userAgent,
		LastIP:       lastIp,
		hub:          hub,
		done:         make(chan struct{}),
		isActive:     true,
		lastActivity: time.Now(),
	}
}

func (c *Client) Done() <-chan struct{} {
	return c.done
}

func (c *Client) IsActive() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.isActive
}

func (c *Client) Close() {

	slog.Debug("Close connection",
		"client_id", c.ID,
		"client_uuid", c.UUID,
		"user_agent", c.UserAgent,
		"remote_addr", c.LastIP,
	)

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.isActive {
		c.isActive = false
		close(c.done)
		close(c.Send)
		c.Conn.Close()
	}
	c.hub.Unregister <- c
}

func (c *Client) ReadPump() {
	defer func() {
		slog.Info("ReadPump stopped",
			"client_id", c.ID,
			"client_uuid", c.UUID)
		c.Close()
	}()

	slog.Debug("ReadPump started",
		"client_id", c.ID,
		"client_uuid", c.UUID,
		"user_agent", c.UserAgent,
		"remote_addr", c.LastIP)

	c.Conn.SetReadLimit(c.hub.config.MaxMessageSize)
	c.Conn.SetReadDeadline(time.Now().Add(c.hub.config.PongWait))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(c.hub.config.PongWait))
		return nil
	})

	var messageCount int
	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				slog.Warn("Unexpected WebSocket close",
					"client_id", c.ID,
					"error", err,
					"total_messages", messageCount)
			} else {
				slog.Debug("WebSocket connection closed",
					"client_id", c.ID,
					"error", err,
					"total_messages", messageCount)
			}
			break
		}

		messageCount++

		slog.Debug("Message received",
			"client_id", c.ID,
			"message_size", len(message),
			"message_preview", string(c.truncateMessage(message)),
			"message_number", messageCount)

		c.hub.Broadcast <- message
	}
}

// func for truncate message for shortly preview
func (c *Client) truncateMessage(msg []byte) []byte {
	if len(msg) > 100 {
		return append(msg[:100], '.', '.', '.')
	}
	return msg
}

func (c *Client) WritePump() {
	ticker := time.NewTicker(c.hub.config.PingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
		slog.Info("WritePump stopped",
			"client_id", c.ID,
			"client_uuid", c.UUID)
	}()

	slog.Debug("WritePump started",
		"client_id", c.ID,
		"ping_interval", c.hub.config.PingPeriod)

	var sentMessages int
	var consecutivePingFailures int
	maxConsecutivePingFailures := 3

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(c.hub.config.WriteWait))
			if !ok {
				slog.Debug("Send channel closed, sending close message",
					"client_id", c.ID)
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			err := c.Conn.WriteMessage(websocket.TextMessage, message)
			if err != nil {
				slog.Warn("Failed to write message",
					"client_id", c.ID,
					"error", err)
				return
			}

			sentMessages++
			c.updateLastActivity()
			slog.Debug("Message sent to client",
				"client_id", c.ID,
				"message_size", len(message),
				"total_sent", sentMessages)

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(c.hub.config.WriteWait))

			err := c.Conn.WriteMessage(websocket.PingMessage, nil)
			if err != nil {
				consecutivePingFailures++
				slog.Warn("Failed to send ping",
					"client_id", c.ID,
					"error", err,
					"consecutive_failures", consecutivePingFailures)

				if consecutivePingFailures >= maxConsecutivePingFailures {
					slog.Info("Max ping failures reached, closing connection",
						"client_id", c.ID,
						"failures", consecutivePingFailures)
					return
				}
			} else {
				consecutivePingFailures = 0
				c.updateLastActivity()
				slog.Debug("Ping sent successfully",
					"client_id", c.ID,
					"consecutive_failures", consecutivePingFailures)
			}

			c.Conn.SetReadDeadline(time.Now().Add(time.Millisecond * 100))
			_, _, err = c.Conn.ReadMessage()
			if err != nil {
				if !websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
					consecutivePingFailures++
					slog.Warn("Failed to read pong response",
						"client_id", c.ID,
						"error", err,
						"consecutive_failures", consecutivePingFailures)

					if consecutivePingFailures >= maxConsecutivePingFailures {
						return
					}
				}
			}
		}
	}
}

func (c *Client) updateLastActivity() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.lastActivity = time.Now()
}

func (c *Client) GetLastActivity() time.Time {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.lastActivity
}
