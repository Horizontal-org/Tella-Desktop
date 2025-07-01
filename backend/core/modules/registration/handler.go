package registration

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{
		service: service,
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

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(struct {
		Status string `json:"status"`
	}{
		Status: "ok",
	})
}
