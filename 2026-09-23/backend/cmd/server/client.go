package main

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"sync"
	"time"
	"uuid"

	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// you'd normally use this for CORS checks in prod
		return true
	},
}

type Client struct {
	hub  *Hub
	conn *websocket.Conn

	send     chan *OutboundMessage
	username string
	rooms    map[uuid.UUID]bool

	once sync.Once
}

func (c *Client) Close() {
	c.once.Do(func() {
		c.conn.Close()
	})
}

func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, payload, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				slog.Error("error reading websocket", "error", err)
			}
			break
		}

		var msg InboundMessage
		if err := json.Unmarshal(payload, &msg); err != nil {
			slog.Error("invalid json payload", "name", c.username, "error", err)
			continue
		}

		switch msg.Action {
		case ActionJoinRoom:
			if msg.RoomID != uuid.Nil() {
				c.hub.join <- &RoomRequest{client: c, roomID: msg.RoomID}
			}
		case ActionLeaveRoom:
			if msg.RoomID != uuid.Nil() && c.rooms[msg.RoomID] {
				c.hub.leave <- &RoomRequest{client: c, roomID: msg.RoomID}
			}
		case ActionSendMessage:
			if msg.RoomID != uuid.Nil() && msg.Content != "" && c.rooms[msg.RoomID] {
				c.hub.broadcast <- &OutboundMessage{
					RoomID:  msg.RoomID,
					Sender:  c.username,
					Content: msg.Content,
				}
			}
		}
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseGoingAway, "server shutting down"))
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}

			jsonMsg, err := json.Marshal(message)
			if err != nil {
				slog.Error("failed to marshal message", "error", err)
				w.Close()
				return
			}
			w.Write(jsonMsg)

			n := len(c.send)
			for i := 0; i < n; i++ {
				jsonMsg, err := json.Marshal(<-c.send)
				if err != nil {
					slog.Error("failed to marshal message", "error", err)
					continue
				}
				w.Write([]byte{'\n'})
				w.Write(jsonMsg)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
