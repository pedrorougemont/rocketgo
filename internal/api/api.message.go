package api

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/pedrorougemont/rocketgo/internal/store/pgstore"
)

type messagePayload struct {
	Message string `json:"message"`
}

type messageResponse struct {
	ID string `json:"id"`
}

func (h apiHandler) handleCreateRoomMessage(w http.ResponseWriter, r *http.Request) {
	rawRoomID := chi.URLParam(r, "room_id")
	roomID, errorMessage, httpErrorCode := h.isValidRoom(rawRoomID, r.Context())
	if httpErrorCode != 0 {
		http.Error(w, errorMessage, httpErrorCode)
		return
	}

	var body messagePayload
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	messageID, err := h.q.InsertMessage(r.Context(), pgstore.InsertMessageParams{RoomID: roomID, Message: body.Message})
	if err != nil {
		slog.Error("failed to insert message", "error", err)
		http.Error(w, "something went wrong", http.StatusInternalServerError)
	}

	rawMessageID := messageID.String()
	data, _ := json.Marshal(messageResponse{ID: rawMessageID})
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(data)

	go h.notifyRoomSubscribers(
		roomNotification{
			RoomID: rawRoomID,
			Kind:   NotificationTypeMessageCreated,
			Value: NotificationMessageCreated{
				ID:      rawMessageID,
				Message: body.Message,
			}})
}

func (h apiHandler) handleGetRoomMessages(w http.ResponseWriter, r *http.Request)        {}
func (h apiHandler) handleGetRoomMessage(w http.ResponseWriter, r *http.Request)         {}
func (h apiHandler) handleReactToMessage(w http.ResponseWriter, r *http.Request)         {}
func (h apiHandler) handleRemoveReactFromMessage(w http.ResponseWriter, r *http.Request) {}
func (h apiHandler) handleMarkMessageAsAnswered(w http.ResponseWriter, r *http.Request)  {}
