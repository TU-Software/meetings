package main

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"slices"
	"uuid"

	db "github.com/TU-Software/meetings/2026-09-23/gen/db"
)

// writeJSON marshals v as JSON into w, setting Content-Type and the given status code.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		slog.Error("failed to write JSON response", "error", err)
	}
}

// writeError writes a JSON {"error": msg} body with the given status code.
func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

// roomResponse is the JSON shape for a room.
type roomResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// messageResponse is the JSON shape for a message.
type messageResponse struct {
	Type     string `json:"type"`
	ID       string `json:"id"`
	RoomID   string `json:"room_id"`
	RoomName string `json:"room_name"`
	Sender   string `json:"sender"`
	Content  string `json:"content"`
}

// handleListRooms handles GET /api/rooms
func handleListRooms(q *db.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rooms, err := q.ListRooms(r.Context())
		if err != nil {
			slog.Error("ListRooms failed", "error", err)
			writeError(w, http.StatusInternalServerError, "failed to list rooms")
			return
		}

		resp := make([]roomResponse, len(rooms))
		for i, room := range rooms {
			resp[i] = roomResponse{
				ID:   room.ID.String(),
				Name: room.Name,
			}
		}
		writeJSON(w, http.StatusOK, resp)
	}
}

// handleCreateRoom handles POST /api/rooms
func handleCreateRoom(q *db.Queries, hub *Hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Name string `json:"name"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Name == "" {
			writeError(w, http.StatusBadRequest, "request body must include a non-empty \"name\"")
			return
		}

		room, err := q.CreateRoom(r.Context(), body.Name)
		if err != nil {
			slog.Error("CreateRoom failed", "error", err)
			writeError(w, http.StatusInternalServerError, "failed to create room")
			return
		}

		resp := roomResponse{ID: room.ID.String(), Name: room.Name}

		// Notify all connected clients that a new room is available.
		hub.notify <- &OutboundMessage{
			Type: OutboundTypeRoomCreated,
			Room: &RoomPayload{ID: room.ID, Name: room.Name},
		}

		writeJSON(w, http.StatusCreated, resp)
	}
}

// handleGetRoomMessages handles GET /api/rooms/{id}
func handleGetRoomMessages(q *db.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		roomID, err := uuid.Parse(r.PathValue("id"))
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid room id")
			return
		}

		room, err := q.GetRoom(r.Context(), roomID)
		if err != nil {
			slog.Error("GetRoom failed", "error", err)
			writeError(w, http.StatusNotFound, "room not found")
			return
		}

		msgs, err := q.GetRecentMessages(r.Context(), roomID)
		if err != nil {
			slog.Error("GetRecentMessages failed", "error", err)
			writeError(w, http.StatusInternalServerError, "failed to fetch messages")
			return
		}

		// GetRecentMessages returns newest-first; reverse so clients receive oldest-first.
		slices.Reverse(msgs)

		resp := make([]messageResponse, len(msgs))
		for i, msg := range msgs {
			resp[i] = messageResponse{
				Type:     string(OutboundTypeMessage),
				ID:       msg.ID.String(),
				RoomID:   msg.RoomID.String(),
				RoomName: room.Name,
				Sender:   msg.Sender,
				Content:  msg.Content,
			}
		}
		writeJSON(w, http.StatusOK, resp)
	}
}
