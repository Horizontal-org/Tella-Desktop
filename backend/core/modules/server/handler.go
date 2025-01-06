package server

import (
    "net/http"
    "Tella-Desktop/backend/core/modules/registration"
    "Tella-Desktop/backend/core/modules/transfer"
)

type Handler struct {
    mux                *http.ServeMux
    registrationHandler *registration.Handler
    transferHandler     *transfer.Handler
}

func NewHandler(
    mux *http.ServeMux,
    registrationHandler *registration.Handler,
    transferHandler *transfer.Handler,
) *Handler {
    return &Handler{
        mux:                mux,
        registrationHandler: registrationHandler,
        transferHandler:     transferHandler,
    }
}

func (h *Handler) SetupRoutes() {
    h.mux.HandleFunc("/api/localsend/v2/register", h.registrationHandler.HandleRegister)
    h.mux.HandleFunc("/api/localsend/v2/prepare-upload", h.transferHandler.HandlePrepare)
    h.mux.HandleFunc("/api/localsend/v2/upload", h.transferHandler.HandleUpload)
}