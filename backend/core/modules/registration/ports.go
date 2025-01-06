package registration

type Service interface {
    Register(device *Device) error
}