package transfer

import (
	"Tella-Desktop/backend/core/modules/filestore"
	"Tella-Desktop/backend/utils/transferutils"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type Handler struct {
	service       Service
	fileService   filestore.Service
	defaultFolder int64 // Default folder ID to store received files
}

func NewHandler(service Service, fileService filestore.Service, defaultFolder int64) *Handler {
	return &Handler{
		service:       service,
		fileService:   fileService,
		defaultFolder: defaultFolder,
	}
}

func (h *Handler) HandlePrepare(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request PrepareUploadRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		runtime.LogError(r.Context(), "Failed to decode prepare upload request: "+err.Error())
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if err := request.Validate(); err != nil {
		runtime.LogError(r.Context(), "Invalid prepare upload request: "+err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	response, err := h.service.PrepareUpload(&request)
	if err != nil {
		runtime.LogError(r.Context(), "Failed to prepare upload: "+err.Error())
		http.Error(w, "Failed to prepare upload", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		runtime.LogError(r.Context(), "Failed to encode response: "+err.Error())
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func (h *Handler) HandleUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	sessionID := r.URL.Query().Get("sessionId")
	transmissionID := r.URL.Query().Get("transmissionId")
	fileID := r.URL.Query().Get("fileId")

	if err := transferutils.ValidateUploadRequest(sessionID, transmissionID, fileID); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	transfer, err := h.service.GetTransfer(fileID)
	if err != nil {
		fmt.Printf("Transfer not found for fileID: %s\n", fileID)
		http.Error(w, "Transfer not found", http.StatusNotFound)
		return
	}

	fileName := transfer.FileInfo.FileName
	mimeType := transfer.FileInfo.FileType

	fmt.Printf("Receiving file: %s (type: %s)\n", fileName, mimeType)

	if err := h.service.HandleUpload(
		sessionID,
		transmissionID,
		fileID,
		r.Body,
		fileName,
		mimeType,
		h.defaultFolder,
	); err != nil {
		switch err {
		case transferutils.ErrTransferNotFound:
			http.Error(w, "Transfer not found", http.StatusNotFound)
		case transferutils.ErrInvalidSession:
			http.Error(w, "Invalid session", http.StatusUnauthorized)
		case transferutils.ErrTransferComplete:
			http.Error(w, "Transfer already completed", http.StatusConflict)
		default:
			runtime.LogError(r.Context(), "Upload failed: "+err.Error())
			http.Error(w, "Failed to store file", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(UploadResponse{Success: true})
}
