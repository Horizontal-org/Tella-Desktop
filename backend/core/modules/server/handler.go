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

func (h *Handler) SetupRoutes() {
	h.mux.HandleFunc("/api/v1/register", h.registrationHandler.HandleRegister)
	h.mux.HandleFunc("/api/v1/prepare-upload", h.transferHandler.HandlePrepare)
	h.mux.HandleFunc("/api/v1/upload", h.transferHandler.HandleUpload)
}
