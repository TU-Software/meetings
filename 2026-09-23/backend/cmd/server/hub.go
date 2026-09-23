package main

import (
	"context"
	"log/slog"
	"uuid"

	db "github.com/TU-Software/meetings/2026-09-23/gen/db"
)

// Room holds the subscriber set and metadata for a chat room.
type Room struct {
	id          uuid.UUID
	name        string
	subscribers map[*Client]bool
}

// RoomRequest is sent over the join/leave channels.
type RoomRequest struct {
	client *Client
	roomID uuid.UUID
}

type Hub struct {
	clients map[*Client]bool
	rooms   map[uuid.UUID]*Room

	register   chan *Client
	unregister chan *Client
	join       chan *RoomRequest
	leave      chan *RoomRequest
	broadcast  chan *OutboundMessage
	// notify broadcasts a message to every connected client (not room-scoped).
	notify   chan *OutboundMessage
	shutdown chan struct{}

	queries *db.Queries
}

func NewHub(queries *db.Queries) *Hub {
	return &Hub{
		register:   make(chan *Client),
		unregister: make(chan *Client),
		join:       make(chan *RoomRequest),
		leave:      make(chan *RoomRequest),
		broadcast:  make(chan *OutboundMessage),
		notify:     make(chan *OutboundMessage),
		shutdown:   make(chan struct{}),
		clients:    make(map[*Client]bool),
		rooms:      make(map[uuid.UUID]*Room),
		queries:    queries,
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true

		case client := <-h.unregister:
			h.removeClient(client)

		case req := <-h.join:
			h.addClientToRoom(req.client, req.roomID)

		case req := <-h.leave:
			h.removeClientFromRoom(req.client, req.roomID)

		case message := <-h.broadcast:
			h.persistAndBroadcast(message)

		case msg := <-h.notify:
			for client := range h.clients {
				select {
				case client.send <- msg:
				default:
					h.removeClient(client)
				}
			}

		case <-h.shutdown:
			for client := range h.clients {
				h.removeClient(client)
			}
			return
		}
	}
}

func (h *Hub) removeClient(client *Client) {
	if _, ok := h.clients[client]; ok {
		for roomID := range client.rooms {
			h.removeClientFromRoom(client, roomID)
		}
		delete(h.clients, client)
		close(client.send)
	}
}

func (h *Hub) Shutdown() {
	close(h.shutdown)
}

func (h *Hub) addClientToRoom(client *Client, roomID uuid.UUID) {
	room, ok := h.rooms[roomID]
	if !ok {
		// First client joining this room in the current session: fetch from DB.
		row, err := h.queries.GetRoom(context.Background(), roomID)
		if err != nil {
			slog.Error("join failed: room not found", "room_id", roomID, "error", err)
			return
		}
		room = &Room{
			id:          row.ID,
			name:        row.Name,
			subscribers: make(map[*Client]bool),
		}
		h.rooms[roomID] = room
	}

	if room.subscribers[client] {
		return
	}
	room.subscribers[client] = true
	client.rooms[roomID] = true

}

func (h *Hub) removeClientFromRoom(client *Client, roomID uuid.UUID) {
	room, ok := h.rooms[roomID]
	if !ok {
		return
	}
	if !room.subscribers[client] {
		return
	}

	delete(room.subscribers, client)
	delete(client.rooms, roomID)

	if len(room.subscribers) == 0 {
		delete(h.rooms, roomID)
		return
	}

}

func (h *Hub) persistAndBroadcast(message *OutboundMessage) {
	room, ok := h.rooms[message.RoomID]
	if !ok {
		slog.Warn("dropping message: unknown room", "room_id", message.RoomID)
		return
	}
	row, err := h.queries.CreateMessage(context.Background(), db.CreateMessageParams{
		RoomID:  message.RoomID,
		Sender:  message.Sender,
		Content: message.Content,
	})
	if err != nil {
		slog.Error("failed to persist message — not broadcasting", "room", room.name, "error", err)
		return
	}
	message.Type = OutboundTypeMessage
	message.ID = row.ID
	message.RoomName = room.name
	h.broadcastToRoom(room, message)
}

func (h *Hub) broadcastToRoom(room *Room, msg *OutboundMessage) {
	for client := range room.subscribers {
		if h.clients[client] {
			select {
			case client.send <- msg:
			default:
				h.removeClient(client)
			}
		}
	}
}
