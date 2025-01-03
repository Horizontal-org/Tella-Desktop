package models

type Device struct {
    Alias       string
    Version     string
    DeviceModel string
    DeviceType  string
    Fingerprint string
    Port        int
    Protocol    string
    Download    bool
}