package api

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type roomPayload struct {
	Theme string `json:"theme"`
}

type roomResponse struct {
	ID string `json:"id"`
}

func (h apiHandler) isValidRoom(rawRoomID string, ctx context.Context) (uuid.UUID, string, int) {
	roomID, err := uuid.Parse(rawRoomID)
	if err != nil {
		return uuid.Nil, "room id is not well-formed", http.StatusBadRequest
	}

	_, err = h.q.GetRoom(ctx, roomID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return uuid.Nil, "room not found", http.StatusBadRequest
		}

		return uuid.Nil, "something went wrong", http.StatusInternalServerError
	}

	return roomID, "", 0
}

func (h apiHandler) handleCreateRoom(w http.ResponseWriter, r *http.Request) {
	var body roomPayload
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid json payload", http.StatusBadRequest)
		return
	}

	_, err := h.q.GetRoomByTheme(r.Context(), body.Theme)
	if err == nil {
		slog.Warn("theme already exists", "error", err)
		http.Error(w, "theme already exists", http.StatusBadRequest)
		return
	}

	if !errors.Is(err, pgx.ErrNoRows) {
		slog.Error("failed to verify if the room exists in the db", "error", err)
		http.Error(w, "something went wrong", http.StatusInternalServerError)
		return
	}

	roomID, err := h.q.InsertRoom(r.Context(), body.Theme)
	if err != nil {
		slog.Error("failed to insert room", "error", err)
		http.Error(w, "something went wrong", http.StatusInternalServerError)
		return
	}

	data, _ := json.Marshal(roomResponse{ID: roomID.String()})
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(data)
}

func (h apiHandler) handleGetRooms(w http.ResponseWriter, r *http.Request) {}
