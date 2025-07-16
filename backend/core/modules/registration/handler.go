package registration

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type Handler struct {
	service Service
	ctx     context.Context
}

func NewHandler(service Service, ctx context.Context) *Handler {
	return &Handler{
		service: service,
		ctx:     ctx,
	}
}

func (h *Handler) HandleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	requestBody, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}

	fmt.Printf("Raw registration request body: %s\n", string(requestBody))

	var request struct {
		PIN   string `json:"pin"`
		Nonce string `json:"nonce"`
	}

	if err := json.Unmarshal(requestBody, &request); err != nil {
		fmt.Printf("Error decoding request: %v\n", err)
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	if request.PIN == "" || request.Nonce == "" {
		http.Error(w, "Missing required parameters", http.StatusBadRequest)
		return
	}

	sessionID, err := h.service.CreateSession(request.PIN, request.Nonce)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(struct {
		SessionID string `json:"sessionId"`
	}{
		SessionID: sessionID,
	})
}

func (h *Handler) HandlePing(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	runtime.EventsEmit(h.ctx, "ping-received", map[string]interface{}{
		"timestamp": "now",
		"message":   "Device attempting to connect",
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(struct {
		Status string `json:"status"`
	}{
		Status: "ok",
	})
}
