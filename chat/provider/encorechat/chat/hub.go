package chat

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/cockroachdb/errors"
	"github.com/gorilla/websocket"

	"encore.dev/rlog"
)

// Example copied and adapted from
// https://github.com/gorilla/websocket/tree/main/examples/chat
func NewHub(ctx context.Context, msgHandler func(ctx context.Context, channel, author string, content []byte) error) *Hub {
	hub := &Hub{
		broadcast:  make(chan *channelMessage),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[string]map[*Client]bool),
		msgHandler: msgHandler,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
	}
	go hub.run(ctx)
	return hub
}

type Hub struct {
	upgrader websocket.Upgrader

	// Registered clients.
	clients map[string]map[*Client]bool

	// Inbound messages from the clients.
	broadcast chan *channelMessage

	// Register requests from the clients.
	register chan *Client

	// Unregister requests from clients.
	unregister chan *Client

	msgHandler func(ctx context.Context, channel, author string, content []byte) error

	ctx context.Context
}

type channelMessage struct {
	ChannelID string `json:"channel_id"`
	UserID    string `json:"user_id"`
	Content   string `json:"content"`
}

func (h *Hub) BroadCast(ctx context.Context, channelID, userID, content string) {
	h.broadcast <- &channelMessage{
		ChannelID: channelID,
		UserID:    userID,
		Content:   content,
	}
}

func (h *Hub) Subscribe(channelID, userID string, w http.ResponseWriter, r *http.Request) error {
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return errors.Wrap(err, "upgrade connection")
	}
	client := &Client{
		hub:       h,
		conn:      conn,
		send:      make(chan []byte, 256),
		channelID: channelID,
		userID:    userID,
	}
	h.register <- client

	// Allow collection of memory referenced by the caller by doing all work in
	// new goroutines.
	go client.writePump(h.ctx)
	go client.readPump(h.ctx)
	return nil
}

func (h *Hub) run(ctx context.Context) {
	h.ctx = ctx
	for {
		select {
		case <-ctx.Done():
			return
		case client := <-h.register:
			if h.clients[client.channelID] == nil {
				h.clients[client.channelID] = make(map[*Client]bool)
			}
			h.clients[client.channelID][client] = true
		case client := <-h.unregister:
			if _, ok := h.clients[client.channelID][client]; ok {
				delete(h.clients[client.channelID], client)
				close(client.send)
			}
		case message := <-h.broadcast:
			if h.clients[message.ChannelID] == nil {
				continue
			}
			msgData, err := json.Marshal(message)
			if err != nil {
				rlog.Error("marshal message", "error", err)
				continue
			}
			for client := range h.clients[message.ChannelID] {
				select {
				case client.send <- msgData:
				default:
					close(client.send)
					delete(h.clients[message.ChannelID], client)
				}
			}
		}
	}
}
