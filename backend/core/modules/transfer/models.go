package transfer

type Transfer struct {
    ID            string
    SessionID     string
    Token         string
    FileName      string
    Size          int64
    FileType      string
    Status        string
}

type PrepareUploadRequest struct {
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
    Files map[string]struct {
        ID       string `json:"id"`
        FileName string `json:"fileName"`
        Size     int64  `json:"size"`
        FileType string `json:"fileType"`
    } `json:"files"`
}

type PrepareUploadResponse struct {
    SessionID string            `json:"sessionId"`
    Files    map[string]string `json:"files"`
}
