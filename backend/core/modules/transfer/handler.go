package transfer

import (
    "encoding/json"
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

func (h *Handler) HandlePrepare(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        w.WriteHeader(http.StatusMethodNotAllowed)
        return
    }

    var request PrepareUploadRequest
    if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    response, err := h.service.PrepareUpload(&request)
    if err != nil {
        http.Error(w, "Failed to prepare upload", http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
}

func (h *Handler) HandleUpload(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        w.WriteHeader(http.StatusMethodNotAllowed)
        return
    }

    sessionId := r.URL.Query().Get("sessionId")
    fileId := r.URL.Query().Get("fileId")
    token := r.URL.Query().Get("token")

    if !h.service.ValidateTransfer(sessionId, fileId, token) {
        http.Error(w, "Invalid transfer", http.StatusBadRequest)
        return
    }

    file, header, err := r.FormFile("file")
    if err != nil {
        http.Error(w, "Failed to get file", http.StatusBadRequest)
        return
    }
    defer file.Close()

    if err := h.service.SaveFile(sessionId, fileId, token, header.Filename, file); err != nil {
        http.Error(w, "Failed to save file", http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusOK)
}