package registration

type Service interface {
	CreateSession(pin string, nonce string) (string, error)
	SetPINCode(pinCode string)
}
