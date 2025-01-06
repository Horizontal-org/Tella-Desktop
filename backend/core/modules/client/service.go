// core/modules/client/service.go
package client

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "mime/multipart"
    "net/http"
    "github.com/google/uuid"
    "github.com/wailsapp/wails/v2/pkg/runtime"
)

type service struct {
    ctx context.Context
}

func NewService(ctx context.Context) Service {
    return &service{
        ctx: ctx,
    }
}

func (s *service) RegisterWithDevice(ip string, port int) error {
    client := &http.Client{}

    regRequest := struct {
        Alias       string `json:"alias"`
        Version     string `json:"version"`
        DeviceModel string `json:"deviceModel"`
        DeviceType  string `json:"deviceType"`
        Fingerprint string `json:"fingerprint"`
        Port        int    `json:"port"`
        Protocol    string `json:"protocol"`
        Download    bool   `json:"download"`
    }{
        Alias:       "TellaDesktop",
        Version:     "2.1",   
        DeviceModel: "Desktop",
        DeviceType:  "desktop",
        Fingerprint: uuid.New().String(),
        Port:        port,
        Protocol:    "http",
        Download:    false,
    }

    payload, err := json.Marshal(regRequest)
    if err != nil {
        return fmt.Errorf("failed to marshal registration request: %v", err)
    }

    url := fmt.Sprintf("http://%s:%d/api/localsend/v2/register", ip, port)
    runtime.LogInfo(s.ctx, fmt.Sprintf("Attempting to connect to: %s", url))

    resp, err := client.Post(url, "application/json", bytes.NewBuffer(payload))
    if err != nil {
        return fmt.Errorf("failed to send registration request: %v", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return fmt.Errorf("registration failed with status: %d", resp.StatusCode)
    }

    return nil
}

func (s *service) SendTestFile(ip string, port int, pin string) error {
    client := &http.Client{}
    
    // Prepare upload request
    prepareRequest := struct {
        Info struct {
            Alias       string `json:"alias"`
            Version     string `json:"version"`
            DeviceModel string `json:"deviceModel"`
            DeviceType  string `json:"deviceType"`
            Fingerprint string `json:"fingerprint"`
            Port        int    `json:"port"`
            Protocol    string `json:"protocol"`
            Download    bool   `json:"download"`
        } `json:"info"`
        Files map[string]interface{} `json:"files"`
    }{
        Files: make(map[string]interface{}),
    }

    prepareRequest.Info.Alias = "TellaDesktop"
    prepareRequest.Info.Version = "2.1"
    prepareRequest.Info.DeviceModel = "Desktop"
    prepareRequest.Info.DeviceType = "desktop"
    prepareRequest.Info.Fingerprint = "test-fingerprint"
    prepareRequest.Info.Port = port
    prepareRequest.Info.Protocol = "http"
    prepareRequest.Info.Download = false

    fileId := uuid.New().String()
    prepareRequest.Files[fileId] = map[string]interface{}{
        "id":       fileId,
        "fileName": "test.txt",
        "size":     16,
        "fileType": "text/plain",
    }

    payload, err := json.Marshal(prepareRequest)
    if err != nil {
        return fmt.Errorf("failed to marshal prepare request: %v", err)
    }

    prepareURL := fmt.Sprintf("http://%s:%d/api/localsend/v2/prepare-upload?pin=%s", ip, port, pin)
    resp, err := client.Post(prepareURL, "application/json", bytes.NewBuffer(payload))
    if err != nil {
        return fmt.Errorf("failed to prepare upload: %v", err)
    }

    var prepareResp struct {
        SessionId string            `json:"sessionId"`
        Files    map[string]string `json:"files"`
    }
    if err := json.NewDecoder(resp.Body).Decode(&prepareResp); err != nil {
        resp.Body.Close()
        return fmt.Errorf("failed to decode prepare response: %v", err)
    }
    resp.Body.Close()

    // Get token for our file
    token, ok := prepareResp.Files[fileId]
    if !ok {
        return fmt.Errorf("no token received for file")
    }

    // Upload file
    body := &bytes.Buffer{}
    writer := multipart.NewWriter(body)
    part, err := writer.CreateFormFile("file", "test.txt")
    if err != nil {
        return fmt.Errorf("failed to create form file: %v", err)
    }
    
    if _, err := part.Write([]byte("Hello from Tella Desktop!")); err != nil {
        return fmt.Errorf("failed to write file content: %v", err)
    }
    writer.Close()

    uploadURL := fmt.Sprintf(
        "http://%s:%d/api/localsend/v2/upload?fileId=%s&token=%s&sessionId=%s",
        ip, port, fileId, token, prepareResp.SessionId,
    )

    req, err := http.NewRequest("POST", uploadURL, body)
    if err != nil {
        return fmt.Errorf("failed to create upload request: %v", err)
    }
    req.Header.Set("Content-Type", writer.FormDataContentType())

    resp, err = client.Do(req)
    if err != nil {
        return fmt.Errorf("failed to upload file: %v", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return fmt.Errorf("upload failed with status %d", resp.StatusCode)
    }

    return nil
}