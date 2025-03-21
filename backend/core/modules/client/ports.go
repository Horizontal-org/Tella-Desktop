package client

type Service interface {
	RegisterWithDevice(ip string, port int, pin string) error
	SendTestFile(ip string, port int, pin string) error
}
