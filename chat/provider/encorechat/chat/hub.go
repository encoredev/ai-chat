package chat

import (
	"context"
	"encoding/json"
	"net/http"

	"golang.org/x/exp/maps"

	"github.com/cockroachdb/errors"
	"github.com/gorilla/websocket"

	"encore.app/pkg/fns"
	"encore.dev/rlog"
)

// Example copied and adapted from
// https://github.com/gorilla/websocket/tree/main/examples/chat
func NewHub(ctx context.Context, msgHandler messageHandler) *Hub {
	hub := &Hub{
		broadcast:  make(chan *ClientMessage),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
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

type messageHandler func(ctx context.Context, clientMessage *ClientMessage) error

type Hub struct {
	upgrader websocket.Upgrader

	// Registered clients.
	clients map[*Client]bool

	// Inbound messages from the clients.
	broadcast chan *ClientMessage

	// Register requests from the clients.
	register chan *Client

	// Unregister requests from clients.
	unregister chan *Client

	msgHandler messageHandler

	ctx context.Context
}

func (h *Hub) BroadCast(ctx context.Context, msg *ClientMessage) {
	h.broadcast <- msg
}

func (h *Hub) Subscribe(w http.ResponseWriter, r *http.Request) error {
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return errors.Wrap(err, "upgrade connection")
	}
	client := &Client{
		hub:      h,
		conn:     conn,
		send:     make(chan []byte, 256),
		channels: map[string]bool{},
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
			h.clients[client] = true
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
		case message := <-h.broadcast:
			clients := fns.FilterParam(maps.Keys(h.clients), message.ConversationId, (*Client).SubscribedToChannel)
			if len(clients) == 0 {
				continue
			}
			message.Type = "message"
			msgData, err := json.Marshal(message)
			if err != nil {
				rlog.Error("marshal message", "error", err)
				continue
			}
			for _, client := range clients {
				select {
				case client.send <- msgData:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
		}
	}
}
