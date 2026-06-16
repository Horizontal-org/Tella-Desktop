package server

import (
	"Tella-Desktop/backend/core/modules/registration"
	"Tella-Desktop/backend/core/modules/transfer"
	"net/http"
)

type Handler struct {
	mux                 *http.ServeMux
	registrationHandler *registration.Handler
	transferHandler     *transfer.Handler
}

func NewHandler(
	mux *http.ServeMux,
	registrationHandler *registration.Handler,
	transferHandler *transfer.Handler,
) *Handler {
	return &Handler{
		mux:                 mux,
		registrationHandler: registrationHandler,
		transferHandler:     transferHandler,
	}
}

// TODO (2026-06-15): implement /api/v2/* 
// TODO (2026-06-15): keep /api/v1/ping /api/v1/register around and serve legacy responses to sent queries there
func (h *Handler) SetupRoutes(pinFingerprint func (string) error) {
	h.mux.HandleFunc("/api/v1/ping", h.registrationHandler.HandlePing)
	h.mux.HandleFunc("/api/v1/register", func(w http.ResponseWriter, r *http.Request) {
		h.registrationHandler.HandleRegister(w, r, pinFingerprint)
	})
	h.mux.HandleFunc("/api/v1/prepare-upload", h.transferHandler.HandlePrepare)
	h.mux.HandleFunc("/api/v1/upload", h.transferHandler.HandleUpload)
	h.mux.HandleFunc("/api/v1/close-connection", h.transferHandler.HandleCloseConnection)
	h.mux.HandleFunc("/", func (res http.ResponseWriter, req *http.Request) {
		http.Error(res, "Not found", http.StatusNotFound)
	})
}
