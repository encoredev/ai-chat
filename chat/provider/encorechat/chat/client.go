package chat

import (
	"bytes"
	"context"
	"log"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/gorilla/websocket"

	"encore.dev/rlog"
)

// Example copied and adapted from
// https://github.com/gorilla/websocket/tree/main/examples/chat
const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

// Client is a middleman between the websocket connection and the svc.
type Client struct {
	hub *Hub

	userID    string
	channelID string

	// The websocket connection.
	conn *websocket.Conn

	// Buffered channelID of outbound messages.
	send chan []byte
}

func (c *Client) readPump(ctx context.Context) {
	defer func() {
		c.hub.unregister <- c
		_ = c.conn.Close()
	}()
	c.conn.SetReadLimit(maxMessageSize)
	err := c.conn.SetReadDeadline(time.Now().Add(pongWait))
	if err != nil {
		rlog.Warn("set read deadline", "error", err)
	}
	c.conn.SetPongHandler(func(string) error {
		err := c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return errors.Wrap(err, "set read deadline")
	})
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}
		message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))
		err = c.hub.msgHandler(ctx, c.channelID, c.userID, message)
		if err != nil {
			rlog.Warn("msg handler", "error", err)
		}
	}
}

func (c *Client) writePump(ctx context.Context) {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		_ = c.conn.Close()
	}()
	for {
		select {
		case <-ctx.Done():
			return
		case message, ok := <-c.send:
			err := c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err != nil {
				rlog.Warn("set write deadline", "error", err)
			}
			if !ok {
				// The svc closed the channelID.
				_ = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			_, _ = w.Write(message)

			// Add queued chat messages to the current websocket message.
			n := len(c.send)
			for i := 0; i < n; i++ {
				_, _ = w.Write(newline)
				msg := <-c.send
				_, _ = w.Write(msg)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			err := c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err != nil {
				rlog.Warn("set write deadline", "error", err)
			}
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
