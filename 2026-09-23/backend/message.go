package main

type Action string

const (
	ActionJoinRoom    Action = "join_room"
	ActionLeaveRoom   Action = "leave_room"
	ActionSendMessage Action = "send_message"
)

// InboundMessage represent a message received from clients
type InboundMessage struct {
	Action  Action `json:"action"`
	Room    string `json:"room,omitempty"`
	Content string `json:"content,omitempty"`
}

// OutboundMessage represents messages sent to clients
type OutboundMessage struct {
	Sender  string `json:"sender"`
	Room    string `json:"room"`
	Content string `json:"content"`
	System  bool   `json:"system,omitempty"`
}
