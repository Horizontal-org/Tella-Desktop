// this is a test moodule to test the protocol from the desktop app. Should be removed in production
package client

import (
	"bytes"
	"context"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type service struct {
	ctx       context.Context
	client    *http.Client
	sessionID string
}

func NewService(ctx context.Context) Service {
	// Create a custom transport that skips certificate verification
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true, // Required for self-signed certificates
			VerifyPeerCertificate: func(rawCerts [][]byte, verifiedChains [][]*x509.Certificate) error {
				if len(rawCerts) > 0 {
					// Compute hash of the certificate
					hash := sha256.Sum256(rawCerts[0])
					hashStr := hex.EncodeToString(hash[:])
					fmt.Printf("Client Received Certificate Hash: %s\n", hashStr)

					runtime.EventsEmit(ctx, "client-certificate-hash", hashStr)
				}
				return nil
			},
		},
	}

	// Create HTTP client with custom transport
	client := &http.Client{
		Transport: transport,
	}

	return &service{
		ctx:    ctx,
		client: client,
	}
}

func (s *service) RegisterWithDevice(ip string, port int, pin string) error {
	regRequest := struct {
		PIN   string `json:"pin"`
		Nonce string `json:"nonce"`
	}{
		PIN:   pin,
		Nonce: uuid.New().String(),
	}

	payload, err := json.Marshal(regRequest)
	if err != nil {
		return fmt.Errorf("failed to marshal registration request: %v", err)
	}

	url := fmt.Sprintf("https://%s:%d/api/v1/register", ip, port)
	runtime.LogInfo(s.ctx, fmt.Sprintf("Attempting to connect to: %s", url))

	resp, err := s.client.Post(url, "application/json", bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("failed to send registration request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("registration failed with status: %d", resp.StatusCode)
	}

	var response struct {
		SessionID string `json:"session_id"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return fmt.Errorf("failed to decode registration response: %v", err)
	}

	s.sessionID = response.SessionID
	runtime.LogInfo(s.ctx, fmt.Sprintf("Successfully registered with session ID: %s", s.sessionID))
	return nil
}

func (s *service) SendTestFile(ip string, port int, pin string) error {
	if s.sessionID == "" {
		return fmt.Errorf("not registered with device, please register first")
	}

	// Create a test file with unique ID
	fileID := uuid.New().String()
	fileContent := []byte("Hello from Tella Desktop!")
	fileName := "test.txt"
	fileHash := sha256.Sum256(fileContent)

	// Prepare upload request
	prepareRequest := struct {
		Title     string `json:"title"`
		SessionID string `json:"sessionId"`
		Metadata  struct {
			Files map[string]struct {
				ID       string `json:"id"`
				FileName string `json:"fileName"`
				Size     int64  `json:"size"`
				FileType string `json:"fileType"`
				SHA256   string `json:"sha256"`
			} `json:"files"`
		} `json:"metadata"`
	}{
		Title:     "Test Upload",
		SessionID: s.sessionID,
	}

	prepareRequest.Metadata.Files = make(map[string]struct {
		ID       string `json:"id"`
		FileName string `json:"fileName"`
		Size     int64  `json:"size"`
		FileType string `json:"fileType"`
		SHA256   string `json:"sha256"`
	})

	prepareRequest.Metadata.Files[fileID] = struct {
		ID       string `json:"id"`
		FileName string `json:"fileName"`
		Size     int64  `json:"size"`
		FileType string `json:"fileType"`
		SHA256   string `json:"sha256"`
	}{
		ID:       fileID,
		FileName: fileName,
		Size:     int64(len(fileContent)),
		FileType: "text/plain",
		SHA256:   hex.EncodeToString(fileHash[:]),
	}

	// Send prepare request
	prepareURL := fmt.Sprintf("https://%s:%d/api/v1/prepare-upload", ip, port)
	preparePayload, err := json.Marshal(prepareRequest)
	if err != nil {
		return fmt.Errorf("failed to marshal prepare request: %v", err)
	}

	prepareResp, err := s.client.Post(prepareURL, "application/json", bytes.NewBuffer(preparePayload))
	if err != nil {
		return fmt.Errorf("failed to send prepare request: %v", err)
	}
	defer prepareResp.Body.Close()

	if prepareResp.StatusCode != http.StatusOK {
		return fmt.Errorf("prepare request failed with status: %d", prepareResp.StatusCode)
	}

	var prepareResponse struct {
		TransmissionID string `json:"transmissionId"`
	}
	if err := json.NewDecoder(prepareResp.Body).Decode(&prepareResponse); err != nil {
		return fmt.Errorf("failed to decode prepare response: %v", err)
	}

	// Create multipart form data for file upload
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", fileName)
	if err != nil {
		return fmt.Errorf("failed to create form file: %v", err)
	}

	if _, err := part.Write(fileContent); err != nil {
		return fmt.Errorf("failed to write file content: %v", err)
	}
	writer.Close()

	// Send file upload request
	uploadURL := fmt.Sprintf(
		"https://%s:%d/api/v1/upload?sessionId=%s&transmissionId=%s&fileId=%s",
		ip, port, s.sessionID, prepareResponse.TransmissionID, fileID,
	)

	uploadReq, err := http.NewRequest("POST", uploadURL, body)
	if err != nil {
		return fmt.Errorf("failed to create upload request: %v", err)
	}
	uploadReq.Header.Set("Content-Type", writer.FormDataContentType())

	uploadResp, err := s.client.Do(uploadReq)
	if err != nil {
		return fmt.Errorf("failed to upload file: %v", err)
	}
	defer uploadResp.Body.Close()

	if uploadResp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(uploadResp.Body)
		return fmt.Errorf("upload failed with status %d: %s", uploadResp.StatusCode, strings.TrimSpace(string(bodyBytes)))
	}

	runtime.LogInfo(s.ctx, "Test file sent successfully")
	return nil
}
