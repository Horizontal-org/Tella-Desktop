package handlers

import (
    "encoding/json"
    "net/http"
    "Tella-Desktop/backend/core/models"
    "Tella-Desktop/backend/core/ports"
)

type UploadHandler struct {
    fileService ports.FileService
}

func NewUploadHandler(fileService ports.FileService) *UploadHandler {
    return &UploadHandler{
        fileService: fileService,
    }
}

func (h *UploadHandler) HandlePrepare(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        w.WriteHeader(http.StatusMethodNotAllowed)
        return
    }

    var request models.PrepareUploadRequest
    if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    response, err := h.fileService.PrepareUpload(&request)
    if err != nil {
        http.Error(w, "Failed to prepare upload", http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
}

func (h *UploadHandler) HandleUpload(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        w.WriteHeader(http.StatusMethodNotAllowed)
        return
    }

    sessionId := r.URL.Query().Get("sessionId")
    fileId := r.URL.Query().Get("fileId")
    token := r.URL.Query().Get("token")

    if !h.fileService.ValidateTransfer(sessionId, fileId, token) {
        http.Error(w, "Invalid transfer", http.StatusBadRequest)
        return
    }

    file, header, err := r.FormFile("file")
    if err != nil {
        http.Error(w, "Failed to get file", http.StatusBadRequest)
        return
    }
    defer file.Close()

    if err := h.fileService.SaveFile(sessionId, fileId, token, header.Filename, file); err != nil {
        http.Error(w, "Failed to save file", http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusOK)
}