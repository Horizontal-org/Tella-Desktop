package registration

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"Tella-Desktop/backend/utils/devlog"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

var log = devlog.Logger("registration")

type PendingRegistration struct {
	PIN      string
	Nonce    string
	Response chan *RegistrationResponse
	Error    chan error
	Created  time.Time
}

type RegistrationResponse struct {
	SessionID string `json:"sessionId"`
}

type Handler struct {
	service             Service
	ctx                 context.Context
	pendingRegistration *PendingRegistration
	mu                  sync.RWMutex
}

func NewHandler(service Service, ctx context.Context) *Handler {
	return &Handler{
		service: service,
		ctx:     ctx,
	}
}

func (h *Handler) HandlePing(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	runtime.EventsEmit(h.ctx, "ping-received", map[string]interface{}{
		"timestamp": time.Now().Unix(),
		"message":   "Device attempting to connect",
		"state":     "waiting",
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(struct {
		Status string `json:"status"`
	}{
		Status: "ok",
	})
}

// rememberClientFingerprint changes tls config of package server. this change also restarts the https server instance.
func (h *Handler) HandleRegister(w http.ResponseWriter, r *http.Request, rememberClientFingerprint func (string) error) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	requestBody, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}

	log("Raw registration request body: %s\n", string(requestBody))

	var request struct {
		PIN   string `json:"pin"`
		Nonce string `json:"nonce"`
		SenderCertificateHash string `json:"senderCertificateHash"`
	}

	if err := json.Unmarshal(requestBody, &request); err != nil {
		log("Error decoding request: %v\n", err)
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	if request.PIN == "" || request.Nonce == "" || len(request.SenderCertificateHash) != 64 {
		http.Error(w, "Missing required parameters", http.StatusBadRequest)
		return
	}

	// Store the pending registration
	h.mu.Lock()
	h.pendingRegistration = &PendingRegistration{
		PIN:      request.PIN,
		Nonce:    request.Nonce,
		Response: make(chan *RegistrationResponse, 1),
		Error:    make(chan error, 1),
		Created:  time.Now(),
	}
	h.mu.Unlock()

	runtime.EventsEmit(h.ctx, "register-request-received", map[string]interface{}{
		"timestamp": time.Now().Unix(),
		"message":   "Sender is requesting to register",
		"state":     "confirm",
	})

	// Wait for user confirmation or timeout
	select {
	case response := <-h.pendingRegistration.Response:
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)

		h.mu.Lock()
		h.pendingRegistration = nil
		// if we're sending a successful response it was because the pin was valid!
		// since PIN was valid, we can now save the sender's certificate hash and restart the server
		// note: since `rememberClientFingerprint` requires a restart of the https server, we likely have to send the
		// response before restarting?
		err = rememberClientFingerprint(request.SenderCertificateHash)
		// TODO cblgh(2026-03-15): figure out something less catastrophic for the app than just panic here. but https
		// handler is a hard place to recover from the death of the https server ^^'
		//
		// maybe emit some kind of event to frontend to signal catastrophic failure && need to restart?
		// runtime.EventsEmit(h.ctx, "register-request-received", map[string]interface{}{
		if err != nil {
			panic(err)
		}
		h.mu.Unlock()

	case err := <-h.pendingRegistration.Error:
		http.Error(w, err.Error(), http.StatusUnauthorized)

		h.mu.Lock()
		h.pendingRegistration = nil
		h.mu.Unlock()

	case <-time.After(30 * time.Second):
		http.Error(w, "Registration timeout", http.StatusRequestTimeout)

		h.mu.Lock()
		h.pendingRegistration = nil
		h.mu.Unlock()
	}
}

func (h *Handler) ConfirmRegistration() error {
	h.mu.RLock()
	pending := h.pendingRegistration
	h.mu.RUnlock()

	if pending == nil {
		return fmt.Errorf("no pending registration to confirm")
	}

	sessionID, err := h.service.CreateSession(pending.PIN, pending.Nonce)
	if err != nil {
		select {
		case pending.Error <- err:
		default:
		}
		return err
	}

	response := &RegistrationResponse{
		SessionID: sessionID,
	}

	select {
	case pending.Response <- response:
		return nil
	default:
		return fmt.Errorf("failed to send registration response")
	}
}

func (h *Handler) RejectRegistration() error {
	h.mu.RLock()
	pending := h.pendingRegistration
	h.mu.RUnlock()

	if pending == nil {
		return fmt.Errorf("no pending registration to reject")
	}

	select {
	case pending.Error <- fmt.Errorf("registration rejected by user"):
		return nil
	default:
		return fmt.Errorf("failed to send rejection")
	}
}
