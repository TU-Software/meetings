package main

import (
	"fmt"
	"sync"
)

type RoomRequest struct {
	client *Client
	room   string
}

type Hub struct {
	clients map[*Client]bool
	rooms   map[string]map[*Client]bool

	register   chan *Client
	unregister chan *Client

	join      chan *RoomRequest
	leave     chan *RoomRequest
	broadcast chan *OutboundMessage

	shutdown chan struct{}
	mu       sync.Mutex
}

func NewHub() *Hub {
	return &Hub{
		register:   make(chan *Client),
		unregister: make(chan *Client),
		join:       make(chan *RoomRequest),
		leave:      make(chan *RoomRequest),
		broadcast:  make(chan *OutboundMessage),
		shutdown:   make(chan struct{}),
		clients:    make(map[*Client]bool),
		rooms:      make(map[string]map[*Client]bool),
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
			h.addClientToRoom(req.client, req.room)

		case req := <-h.leave:
			h.removeClientFromRoom(req.client, req.room)

		case message := <-h.broadcast:
			if subscribers, ok := h.rooms[message.Room]; ok {
				for client := range subscribers {
					if h.clients[client] {
						select {
						case client.send <- message:
						default:
							h.removeClient(client)
						}
					}
				}
			}

		case <-h.shutdown:
			// Close all connected client channels cleanly
			for client := range h.clients {
				h.removeClient(client)
			}
			return
		}
	}
}

func (h *Hub) removeClient(client *Client) {
	if _, ok := h.clients[client]; ok {
		for room := range client.rooms {
			h.removeClientFromRoom(client, room)
		}
		delete(h.clients, client)
		close(client.send) // Safe: deleted from map first
	}
}

func (h *Hub) Shutdown() {
	close(h.shutdown)
}

func (h *Hub) addClientToRoom(client *Client, room string) {
	if h.rooms[room] == nil {
		h.rooms[room] = make(map[*Client]bool)
	}

	if !h.rooms[room][client] {
		h.rooms[room][client] = true
		client.rooms[room] = true

		sysMsg := OutboundMessage{
			Sender:  "System",
			Room:    room,
			Content: fmt.Sprintf("%s joined the room", client.username),
			System:  true,
		}
		h.broadcastToRoom(room, &sysMsg)
	}
}

func (h *Hub) removeClientFromRoom(client *Client, room string) {
	if subscribers, ok := h.rooms[room]; ok {
		if subscribers[client] {
			delete(subscribers, client)
			delete(client.rooms, room)

			if len(subscribers) == 0 {
				delete(h.rooms, room)
			} else {
				sysMsg := &OutboundMessage{
					Sender:  "System",
					Room:    room,
					Content: fmt.Sprintf("%s left the room", client.username),
					System:  true,
				}
				h.broadcastToRoom(room, sysMsg)
			}
		}
	}
}

func (h *Hub) broadcastToRoom(room string, msg *OutboundMessage) {
	if subscribers, ok := h.rooms[room]; ok {
		for client := range subscribers {
			select {
			case client.send <- msg:
			default:
				h.removeClient(client)
			}
		}
	}
}
