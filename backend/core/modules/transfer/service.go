package transfer

import (
	"Tella-Desktop/backend/utils/transferutils"
	"context"
	"database/sql"
	"fmt"
	"io"
	"sync"
	"time"

	"Tella-Desktop/backend/core/modules/filestore"
	"Tella-Desktop/backend/utils/constants"
	"Tella-Desktop/backend/utils/config"

	"github.com/google/uuid"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type service struct {
	config					 config.Config
	ctx              context.Context
	transfers        sync.Map
	pendingTransfers sync.Map
	fileService      filestore.Service
	db               *sql.DB
	sessionIsValid   func(string) bool
	forgetSession    func(string)
	done             chan struct{}
}

type PendingTransfer struct {
	SessionID    string     `json:"sessionId"`
	Title        string     `json:"title"`
	Files        []FileInfo `json:"files"`
	ResponseChan chan *PrepareUploadResponse
	ErrorChan    chan error
	CreatedAt    time.Time
}

type TransferSession struct {
	SessionID         string
	FolderID          int64
	Title             string
	FileIDs           []string
	SeenTransmissions map[string]bool
	ExpiresAt         time.Time
}

// timeout = 10 hours. We use a long timeout so that our fallback for cleaning up memory does not risk causing issues
// with rare very long duration transfers. 
// 
// If we assume transfer speeds of [1MB/s, 6MB/s], then the chosen window gives us a transfered total payload [36GB, 216GB] in the given 10h window.
const REFRESH_TIMEOUT_MIN = 45   // timeout window allows for transfers between [27GB and 162GB] for speeds [1MB/s, 6MB/s]

func NewService(ctx context.Context, fileSerservice filestore.Service, db *sql.DB, sessionIsValid func(string) bool, forgetSession func(string)) Service {
	conf := config.ReadConfig()
	return &service{
		config: conf,
		ctx:              ctx,
		transfers:        sync.Map{},
		pendingTransfers: sync.Map{},
		fileService:      fileSerservice,
		db:               db,
		sessionIsValid: sessionIsValid,
		forgetSession: forgetSession,
		done: make(chan struct{}),
	}
}

func (s *service) PrepareUpload(request *PrepareUploadRequest) (*PrepareUploadResponse, error) {
	pendingTransfer := &PendingTransfer{
		SessionID:    request.SessionID,
		Title:        request.Title,
		Files:        request.Files,
		ResponseChan: make(chan *PrepareUploadResponse, 1),
		ErrorChan:    make(chan error, 1),
		CreatedAt:    time.Now(),
	}

	_, exists := s.pendingTransfers.Load(request.SessionID)
	if exists {
		return nil, fmt.Errorf("pending transfer already exists for session: %s", request.SessionID)
	}

	// correctly checks that the sessionID from the registration is the same as the sessionID arriving in our prepare-upload request
	if !s.sessionIsValid(request.SessionID) {
		return nil, transferutils.ErrInvalidSession
	}

	/* check desktop-defined limits */
	// ensure config.MaxFileCount is not exceeded in prepare-upload data
	if len(request.Files) > s.config.MaxFileCount {
		return nil, transferutils.ErrTransferTooLarge 
	}
	// ensure config.MaxFileFileSizeBytes is not exceeded for any file in prepare-upload data
	for _, file := range request.Files {
		if file.Size > s.config.MaxFileSizeBytes {
			return nil, transferutils.ErrTransferTooLarge 
		}
	}

	s.pendingTransfers.Store(request.SessionID, pendingTransfer)

	runtime.EventsEmit(s.ctx, "prepare-upload-request", map[string]interface{}{
		"sessionId":  request.SessionID,
		"title":      request.Title,
		"files":      request.Files,
		"totalFiles": len(request.Files),
		"transferredFiles": 0,
		"totalSize":  s.calculateTotalSize(request.Files),
	})

	// Cleanup of s.pendingTransfers: select waits until one of the channels has a communication (a channel send event, in all
	// three cases below). For all paths, we make sure that s.pendingTransfers deletes the corresponding sync.Map entry for
	// request.SessionID.
	select {
	case response := <-pendingTransfer.ResponseChan:
		s.pendingTransfers.Delete(request.SessionID)
		return response, nil
	case err := <-pendingTransfer.ErrorChan:
		s.pendingTransfers.Delete(request.SessionID)
		return nil, err
	case <-s.done:
		s.pendingTransfers.Delete(request.SessionID)
		return nil, fmt.Errorf("request timeout - connection was closed by recipient")
	case <-time.After(5 * time.Minute):
		s.pendingTransfers.Delete(request.SessionID)
		return nil, fmt.Errorf("request timeout - no response from recipient")
	}
}

func (s *service) AcceptTransfer(sessionID string) error {
	value, exists := s.pendingTransfers.Load(sessionID)
	if !exists {
		return fmt.Errorf("no pending transfer found for session: %s", sessionID)
	}

	pendingTransfer, ok := value.(*PendingTransfer)
	if !ok {
		return fmt.Errorf("invalid pending transfer data")
	}

	folderID, err := s.createTransferFolder(pendingTransfer.Title)
	if err != nil {
		return fmt.Errorf("failed to create transfer folder: %w", err)
	}

	var fileIDs[] string 
	var responseFiles []FileTransmissionInfo
	for _, fileInfo := range pendingTransfer.Files {
		transmissionID := uuid.New().String()
		transfer := &Transfer{
			TransmissionID: transmissionID,
			SessionID: sessionID,
			FileInfo:  fileInfo,
			Status:    "pending",
		}
		s.transfers.Store(fileInfo.ID, transfer)
		fileIDs = append(fileIDs, fileInfo.ID)

		responseFiles = append(responseFiles, FileTransmissionInfo{
			ID:             fileInfo.ID,
			TransmissionID: transmissionID,
		})
	}

	transferSession := &TransferSession{
		SessionID: sessionID,
		FolderID:  folderID,
		Title:     pendingTransfer.Title,
		FileIDs:   fileIDs,
		SeenTransmissions: make(map[string]bool),
		ExpiresAt: time.Now().Add(REFRESH_TIMEOUT_MIN * time.Minute),
	}

	s.transfers.Store(sessionID+"_session", transferSession)

	// in the event that the session doesn't conclude properly, this fallback mitigates memory leakage by cleaning up the
	// set s.transfers keys for all fileIDs (+ <sessionID>_session) being stored in this routine
	//
	// TODO cblgh(2026-02-17): add explicit lifecycle 'close' function which would also drain this goroutine (otherwise
	// risk for goroutine leak since it's only cleaned up 10h after starting)
	//
	// note: this is currently taken care of by s.endTransfer, but a more orderly exit would be prefered :)
	go (func(fileIDs []string) {
		// 'done' channel fires when application has been locked -> 
		// exit goroutine and allow GC to cleanup reference to this service
		select {
		case <-s.done:
		case <-time.After(constants.CLEAN_UP_SESSION_TIMEOUT_MIN * time.Minute):
			if s == nil { return }
			for _, fileID := range fileIDs {
				s.ForgetTransfer(fileID)
			}
			s.forgetSession(sessionID)
			fileIDs = []string{""}
		}
	})(fileIDs)

	response := &PrepareUploadResponse{
		Files: responseFiles,
	}

	select {
	case pendingTransfer.ResponseChan <- response:
		runtime.LogInfo(s.ctx, fmt.Sprintf("Transfer accepted for session: %s", sessionID))
		return nil
	default:
		return fmt.Errorf("failed to send acceptance response")
	}
}

func (s *service) RejectTransfer(sessionID string) error {
	value, exists := s.pendingTransfers.Load(sessionID)
	if !exists {
		return fmt.Errorf("no pending transfer found for session: %s", sessionID)
	}

	pendingTransfer, ok := value.(*PendingTransfer)
	if !ok {
		return fmt.Errorf("invalid pending transfer data")
	}

	select {
	case pendingTransfer.ErrorChan <- fmt.Errorf("transfer rejected by recipient"):
		runtime.LogInfo(s.ctx, fmt.Sprintf("Transfer rejected for session: %s", sessionID))
		return nil
	default:
		return fmt.Errorf("failed to send rejection response")
	}
}

func (s *service) GetTransfer(fileID string) (*Transfer, error) {
	if value, ok := s.transfers.Load(fileID); ok {
		if transfers, ok := value.(*Transfer); ok {
			return transfers, nil
		}
	}
	return nil, transferutils.ErrTransferNotFound
}

// ForgetTransfer removes the associated ID and any related sessionID. Returns true if the ID was in the map before being removed.
func (s *service) ForgetTransfer(fileID string) bool {
	v, existed := s.transfers.Load(fileID)
	if existed {
		if transfer, ok := v.(*Transfer); ok {
			s.transfers.Delete(transfer.SessionID + "_session")
		}
	}
	s.transfers.Delete(fileID)
	return existed
}

func (s *service) PreUploadValidation(sessionID, transmissionID, fileID string) error {
	if !s.sessionIsValid(sessionID) {
		return transferutils.ErrInvalidSession
	}

	transfer, err := s.GetTransfer(fileID)
	if err != nil {
		return err
	}

	if transfer.SessionID != sessionID {
		return transferutils.ErrInvalidSession
	}

	if transfer.Status == "completed" {
		return transferutils.ErrTransferComplete
	}

	// verify that file's associated transmissionID is correct
	if transfer.TransmissionID != transmissionID {
		return transferutils.ErrInvalidTransmission
	}
	return nil
}

func (s *service) HandleUpload(sessionID, transmissionID, fileID string, reader io.Reader, fileName string, mimeType string, folderID int64) error {
	err := s.PreUploadValidation(sessionID, transmissionID, fileID)
	if err != nil {
		return err
	}
	transfer, err := s.GetTransfer(fileID)
	if err != nil {
		return err
	}
	actualFolderID := folderID
	var ongoingSession *TransferSession
	if sessionValue, exists := s.transfers.Load(sessionID + "_session"); exists {
		if session, ok := sessionValue.(*TransferSession); ok {
			ongoingSession = session
			// transmission IDs are tied to a single file and rendered invalid after the file has been uploaded
			if _, seen := session.SeenTransmissions[transmissionID]; seen {
				// reject transmission ID reuse
				return transferutils.ErrInvalidTransmission
			}
			session.SeenTransmissions[transmissionID] = true

			// time-based expiry of sessions
			// clean up session keys and return err
			if time.Now().After(session.ExpiresAt) {
				s.ForgetTransfer(fileID)
				s.forgetSession(session.SessionID)
				return transferutils.ErrInvalidSession
			} else {
				// the transfer is still valid and ongoing: refresh the expiry
				session.ExpiresAt = time.Now().Add(REFRESH_TIMEOUT_MIN * time.Minute)

				actualFolderID = session.FolderID
			}
		}
	}

	runtime.EventsEmit(s.ctx, "file-receiving", map[string]interface{}{
		"sessionId": sessionID,
		"fileId":    fileID,
		"fileName":  fileName,
		"fileSize":  transfer.FileInfo.Size,
	})

	// NOTE cblgh(2026-02-19): is desktop's current architecture for transfer's handler + service unnecessarily
	// blocking senders?
	//
	// if the sender's last file has been fully transmitted (all bytes received), the sender will not receive an
	// acknowledgement ({ success: true }) until the file has been properly stored and encrypted -- which can take quite a
	// bit of time for large files. this forces the sender to keep the application open while nothing useful is happening
	// on their side

	// TODO: cblgh/(2026-03-06):
	// check that config.MaxFileCount is not exceeded
	// check that config.MaxFileFileSizeBytes is not exceeded for any file by wrapping reader in e.g. a LimitReader or
	// MaxBytesReader

	// NOTE cblgh(2026-02-19): LimitReader returns an EOF, which signals the end of reading activities for io.ReadAll or
	// io.Copy. this means that we will only read exactly as many bytes as the sender has told us this file size is! if
	// the file size is incorrectly stated, we will not read or save "the full file". coupled with the sha256 hashes we
	// can detect any such issues. 
	// limit reading of the response to the claimed filesize. reading more bytes than the claimed size will return an error
	fmt.Println("fileName is", fileName, "claimed size", transfer.FileInfo.Size)
	// limitReader := io.LimitReader(reader, transfer.FileInfo.Size)


	// NOTE cblgh(2026-02-19): running into a bug when uploading a large file as part of many other files; something to the effect of 
	// what is described here https://github.com/googleapis/google-cloud-go/issues/987
	// and somewhat detailed in https://github.com/golang/go/issues/26338
	metadata, err := s.fileService.StoreFile(actualFolderID, transfer.FileInfo.Size, fileName, mimeType, reader)
	transferFailed := err != nil

	if transferFailed {
		transfer.Status = "failed"

		runtime.EventsEmit(s.ctx, "file-receive-failed", map[string]interface{}{
			"sessionId": sessionID,
			"fileId":    fileID,
			"fileName":  fileName,
			"fileSize":  transfer.FileInfo.Size,
		})
	} else {
		transfer.Status = "completed"
	}
	s.transfers.Store(fileID, transfer)

	// determine whether all files in a given transfer resolved (Status == {failed || completed})
	// -> perform session clean up when this happens
	allTransfersResolved := true
	resolveLoop:
	for _, fid := range ongoingSession.FileIDs {
		if v, exists := s.transfers.Load(fid); exists {
			if transferInfo, ok := v.(*Transfer); ok {
				if transferInfo.Status != "completed" && transferInfo.Status != "failed" {
					allTransfersResolved = false
					break resolveLoop
				}
			}
		}
	}

	// note cblgh(2026-02-16): is there ui jank that may happen if we do this cleanup immediately after the last file has been
	// handled?
	if allTransfersResolved {
		s.endTransfer(sessionID)
	}

	// if we've failed & determined whether any transfers are still pending, then we can ret with the err
	if transferFailed {
		return fmt.Errorf("failed to store file: %w", err)
	}

	runtime.EventsEmit(s.ctx, "file-received", map[string]interface{}{
		"sessionId": sessionID,
		"fileId":    fileID,
		"fileName":  fileName,
		"fileSize":  transfer.FileInfo.Size,
	})

	runtime.LogInfo(s.ctx, fmt.Sprintf("File stored successfully in folder %d. ID: %s, Name: %s", actualFolderID, metadata.UUID, metadata.Name))
	return nil
}

func (s *service) endTransfer(sessionID string) {
	sessionValue, exists := s.transfers.Load(sessionID + "_session")
	if exists {
		if session, ok := sessionValue.(*TransferSession); ok {
			for _, fileID := range session.FileIDs {
				s.ForgetTransfer(fileID)
			}
		}
	}
	// clears entry for map in registration service
	s.forgetSession(sessionID)
	// drain the previous goroutine
	close(s.done)
	// setup a new channel
	s.done = make(chan struct{})
}

func (s *service) StopTransfer(sessionID string) {
	s.endTransfer(sessionID)
}

func (s *service) CloseConnection(sessionID string) error {
	if !s.sessionIsValid(sessionID) {
		return transferutils.ErrInvalidSession
	}
	// TODO cblgh(2026-02-16): other than forget transfer session state, what else should we do on close connection?
	s.endTransfer(sessionID)
	return nil
}

func (s *service) Lock() {
	s.pendingTransfers.Clear()
	s.transfers.Clear()
	// we close the channel -> a closed channel will be received on immediately
	close(s.done)
}

func (s *service) calculateTotalSize(files []FileInfo) int64 {
	var total int64
	for _, file := range files {
		total += file.Size
	}
	return total
}

func (s *service) createTransferFolder(title string) (int64, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return 0, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Create folder with the transfer title
	result, err := tx.Exec(`
		INSERT INTO folders (name, parent_id, created_at, updated_at) 
		VALUES (?, NULL, datetime('now'), datetime('now'))
	`, title)
	if err != nil {
		return 0, fmt.Errorf("failed to create folder: %w", err)
	}

	folderID, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get folder ID: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("failed to commit transaction: %w", err)
	}

	runtime.LogInfo(s.ctx, fmt.Sprintf("Created transfer folder '%s' with ID: %d", title, folderID))
	return folderID, nil
}
