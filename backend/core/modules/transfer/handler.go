package transfer

import (
	"Tella-Desktop/backend/core/modules/filestore"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/google/uuid"
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

	// Determine MIME type
	mimeType := header.Header.Get("Content-Type")
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}

	// Store file in TVault using our file storage service
	metadata, err := h.fileService.StoreFile(h.defaultFolder, header.Filename, mimeType, file)
	if err != nil {
		runtime.LogError(r.Context(), fmt.Sprintf("Failed to store file: %v", err))
		http.Error(w, "Failed to store file", http.StatusInternalServerError)
		return
	}

	// Emit event to notify UI about received file
	// runtime.EventsEmit(r.Context(), "file-received", header.Filename)

	// Mark transfer as completed in the service
	h.service.CompleteTransfer(sessionId, fileId)

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "success",
		"fileId": metadata.UUID,
	})
}

// SaveTemporaryFile is a helper method to save a file to temp directory
// when needed for debugging or if we can't use the TVault storage
func (h *Handler) SaveTemporaryFile(sessionId, fileId, fileName string, reader io.Reader) error {
	// Create a random filename for the temporary file
	tempFileName := uuid.New().String() + "_" + fileName

	// Get temp directory
	uploadsDir := getTempUploadsDirectory()

	// Create the full file path
	filePath := filepath.Join(uploadsDir, tempFileName)

	// Create destination file
	dst, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create temporary file: %w", err)
	}
	defer dst.Close()

	// Copy file data
	if _, err := io.Copy(dst, reader); err != nil {
		return fmt.Errorf("failed to write temporary file: %w", err)
	}

	return nil
}

func getTempUploadsDirectory() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(".", "uploads")
	}
	uploadsDir := filepath.Join(homeDir, "Documents", "TellaUploads")
	os.MkdirAll(uploadsDir, 0755)
	return uploadsDir
}
