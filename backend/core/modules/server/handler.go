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

// TODO (2026-06-15): keep /api/v1/ping /api/v1/register around and serve legacy responses to sent queries there
func (h *Handler) SetupRoutes(pinFingerprint func (string) error, getSenderFingerprintCandidate func () string) {
	h.mux.HandleFunc("/api/v2/ping", h.registrationHandler.HandlePing)
	h.mux.HandleFunc("/api/v2/register", func(res http.ResponseWriter, req *http.Request) {
		h.registrationHandler.HandleRegister(res, req, pinFingerprint, getSenderFingerprintCandidate)
	})
	h.mux.HandleFunc("/api/v2/prepare-upload", h.transferHandler.HandlePrepare)
	h.mux.HandleFunc("/api/v2/upload", h.transferHandler.HandleUpload)
	h.mux.HandleFunc("/api/v2/close-connection", h.transferHandler.HandleCloseConnection)

	// handle v1 legacy routes
	h.mux.HandleFunc("/api/v1/ping", func (res http.ResponseWriter, req *http.Request) {
		// TODO (2026-06-18): collectively decide what to respond to old v1-pings. Will 400 crash clients or is it OK?
		http.Error(res, "Bad request", http.StatusBadRequest)
	})
	h.mux.HandleFunc("/api/v1/register", func(res http.ResponseWriter, req *http.Request) {
		http.Error(res, "Rejected", http.StatusForbidden)
	})

	h.mux.HandleFunc("/", func (res http.ResponseWriter, req *http.Request) {
		http.Error(res, "Not found", http.StatusNotFound)
	})
}
