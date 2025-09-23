package ws

import (
	"time"

	"github.com/gorilla/websocket"
)

type Client struct {
	ID        string
	UUID      string
	Conn      *websocket.Conn
	Send      chan []byte
	UserAgent string
	LastIP    string
	hub       *Hub
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
		ID:        id,
		UUID:      uuid,
		Conn:      conn,
		Send:      make(chan []byte, 256),
		UserAgent: userAgent,
		LastIP:    lastIp,
		hub:       hub,
	}
}

func (c *Client) ReadPump() {
	defer func() {
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(c.hub.config.MaxMessageSize)
	c.Conn.SetReadDeadline(time.Now().Add(c.hub.config.PongWait))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(c.hub.config.PongWait))
		return nil
	})

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			break
		}

		c.hub.Broadcast <- message
	}
}

func (c *Client) WritePump() {
	ticker := time.NewTicker(c.hub.config.PingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(c.hub.config.WriteWait))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(c.hub.config.WriteWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
