package api

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/gorilla/websocket"
)

type roomNotification struct {
	RoomID string `json:"-"`
	Kind   string `json:"kind"`
	Value  any    `json:"value"`
}

const NotificationTypeMessageCreated = "message_created"

type NotificationMessageCreated struct {
	ID      string `json:"id"`
	Message string `json:"message"`
}

func (h apiHandler) handleSubscribeRoom(w http.ResponseWriter, r *http.Request) {
	rawRoomID := chi.URLParam(r, "room_id")
	_, httpErrormessage, httpErrorCode := h.isValidRoom(rawRoomID, r.Context())
	if httpErrorCode != 0 {
		http.Error(w, httpErrormessage, httpErrorCode)
		return
	}

	c, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.Warn("failed to upgrade connection", "error", err)
		http.Error(w, "failed to upgrade to ws connection", http.StatusBadRequest)
		return
	}

	defer c.Close()

	ctx, cancel := context.WithCancel(r.Context())

	h.mu.Lock()
	if _, ok := h.subscribers[rawRoomID]; !ok {
		h.subscribers[rawRoomID] = make(map[*websocket.Conn]context.CancelFunc)
	}
	slog.Info("new client connected", "room_id", rawRoomID, "client_ip", r.RemoteAddr)
	h.subscribers[rawRoomID][c] = cancel
	h.mu.Unlock()

	<-ctx.Done()

	h.mu.Lock()
	delete(h.subscribers[rawRoomID], c)
	h.mu.Unlock()
}

func (h apiHandler) notifyRoomSubscribers(msg roomNotification) {
	h.mu.Lock()
	defer h.mu.Unlock()

	subscribers, ok := h.subscribers[msg.RoomID]
	if !ok || len(subscribers) == 0 {
		return
	}

	for conn, cancel := range subscribers {
		if err := conn.WriteJSON(msg); err != nil {
			slog.Error("failed to send message to client", "error", err)
			cancel()
		}
	}
}
