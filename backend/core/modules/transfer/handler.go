package transfer

import (
	"Tella-Desktop/backend/core/modules/filestore"
	"Tella-Desktop/backend/utils/nonces"
	"Tella-Desktop/backend/utils/transferutils"
	"Tella-Desktop/backend/utils/devlog"
	"encoding/json"
	"errors"
	"net/http"
)

var log = devlog.Logger("transfer")

type Handler struct {
	nonceManager  *nonces.NonceManager
	service       Service
	fileService   filestore.Service
	defaultFolder int64 // Default folder ID to store received files
}

func NewHandler(service Service, fileService filestore.Service, defaultFolder int64, nm *nonces.NonceManager) *Handler {
	return &Handler{
		nonceManager:  nm,
		service:       service,
		fileService:   fileService,
		defaultFolder: defaultFolder,
	}
}

type closeInfo struct {
	SessionID string `json:"sessionId"`
}

// TODO cblgh(2026-02-16): this route, and the methods it calls, is currently a semi-functional stub. It could do with some more work after the
// other audit fixes are taking care of.
func (h *Handler) HandleCloseConnection(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var info closeInfo
	if err := json.NewDecoder(r.Body).Decode(&info); err != nil {
		log("Failed to decode close connection request: %s\n", err.Error())
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if info.SessionID == "" {
		log("Close connection request did not contain sessionID")
		http.Error(w, "No sessionID", http.StatusBadRequest)
		return
	}

	err := h.service.CloseConnection(info.SessionID)
	if err != nil {
		log("Failure for close-connection: %s\n", err.Error())
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]bool{"success": true}); err != nil {
		log("Failed to encode response: %s\n", err.Error())
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	// TODO cblgh(2026-02-16): should "close connection" ultimately also stop the https server?
}

func (h *Handler) HandlePrepare(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request PrepareUploadRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		log("Failed to decode prepare upload request: %s\n", err.Error())
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	// TODO cblgh(2026-03-09): improve prepare-upload request validation logic
	if err := request.Validate(); err != nil {
		log("Invalid request format: %s\n", err.Error())
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	err := h.nonceManager.Add(request.Nonce)
	if err != nil {
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	response, err := h.service.PrepareUpload(&request)
	if err != nil {
		httpErrCode := http.StatusInternalServerError
		errMessage := "Failed to prepare upload"
		if errors.Is(err, transferutils.ErrTransferTooLarge) {
			httpErrCode = http.StatusRequestEntityTooLarge
			errMessage = "Content too large"
		} else if errors.Is(err, transferutils.ErrInvalidSession) {
			// return 401
			httpErrCode = http.StatusUnauthorized
			errMessage = "Invalid session ID"
		}
		log("%s: %s\n", errMessage, err.Error())
		http.Error(w, errMessage, httpErrCode)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log("Failed to encode response: %s\n", err.Error())
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

// HandleUpload has its body wrapped in a MaxBytesReader to limit the amount of bytes that will be read of an incoming
// payload. The limit is configurable in the application config.
func (h *Handler) HandleUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	sessionID := r.URL.Query().Get("sessionId")
	transmissionID := r.URL.Query().Get("transmissionId")
	fileID := r.URL.Query().Get("fileId")
	nonce := r.URL.Query().Get("nonce")

	// TODO cblgh(2026-02-17): pass enough information to ValidateUploadRequest that it can actually perform validation
	// or remove the function entirely (it is basically unused)
	if err := transferutils.ValidateUploadRequest(sessionID, transmissionID, fileID); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	transfer, err := h.service.GetTransfer(fileID)
	if err != nil {
		log("Transfer not found for fileID: %s\n", fileID)
		http.Error(w, "Transfer not found", http.StatusNotFound)
		return
	}

	err = h.nonceManager.Add(nonce)
	if err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	fileName := transfer.FileInfo.FileName
	mimeType := transfer.FileInfo.FileType

	log("Receiving file: %s (type: %s)\n", fileName, mimeType)

	// limit reading from body to at most config.MaxFileSyzeBites
	limitedBody := http.MaxBytesReader(w, r.Body, h.service.GetMaxFileSizeLimit())

	// TODO cblgh(2026-02-16): handle situation where transfer has been stopped & HTTPS server should be terminated
	if err := h.service.HandleUpload(
		sessionID,
		transmissionID,
		fileID,
		limitedBody,
		fileName,
		mimeType,
		h.defaultFolder,
	); err != nil {
		switch err {
		case transferutils.ErrTransferNotFound:
			http.Error(w, "Transfer not found", http.StatusNotFound)
		case transferutils.ErrInvalidSession:
			http.Error(w, "Invalid session", http.StatusUnauthorized)
		case transferutils.ErrInvalidTransmission:
			http.Error(w, "Invalid transmission ID", http.StatusUnauthorized)
		case transferutils.ErrTransferTooLarge:
			http.Error(w, "Content too large", http.StatusRequestEntityTooLarge)
		// TODO cblgh(2026-03-12): implement detection of hdd space running out in FileService in order to return this error
		case transferutils.ErrTransferInsufficentSpace:
			http.Error(w, "Insufficient storage space", http.StatusInsufficientStorage)
		case transferutils.ErrTransferComplete:
			http.Error(w, "Transfer already completed", http.StatusConflict)
		case transferutils.ErrTransferHashMismatch:
			http.Error(w, "File hash mismatch", http.StatusNotAcceptable)
		default:
			log("Upload failed: %s\n", err.Error())
			http.Error(w, "Failed to store file", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(UploadResponse{Success: true})
}
