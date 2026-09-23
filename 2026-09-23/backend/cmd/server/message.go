package main

import "uuid"

type Action string

const (
	ActionJoinRoom    Action = "join_room"
	ActionLeaveRoom   Action = "leave_room"
	ActionSendMessage Action = "send_message"
)

// InboundMessage represents a message received from clients.
// All actions identify the room by UUID.
type InboundMessage struct {
	Action  Action    `json:"action"`
	RoomID  uuid.UUID `json:"room_id"`
	Content string    `json:"content,omitempty"`
}

// OutboundType distinguishes the kind of payload sent to clients.
type OutboundType string

const (
	OutboundTypeMessage     OutboundType = "message"
	OutboundTypeRoomCreated OutboundType = "room_created"
)

// OutboundMessage represents a payload sent to clients over the WebSocket.
// The Type field determines which other fields are populated.
type OutboundMessage struct {
	Type OutboundType `json:"type"`

	// message fields
	ID       uuid.UUID `json:"id,omitempty"`
	RoomID   uuid.UUID `json:"room_id,omitempty"`
	RoomName string    `json:"room_name,omitempty"`
	Sender   string    `json:"sender,omitempty"`
	Content  string    `json:"content,omitempty"`

	// room_created fields
	Room *RoomPayload `json:"room,omitempty"`
}

// RoomPayload is the room data carried in a room_created event.
type RoomPayload struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}
