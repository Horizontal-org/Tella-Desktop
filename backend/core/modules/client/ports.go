package client

type Service interface {
    RegisterWithDevice(ip string, port int) error
    SendTestFile(ip string, port int, pin string) error
}