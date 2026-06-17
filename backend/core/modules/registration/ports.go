package registration

type Service interface {
	IsAuthorised(pin, nonce string) (bool, error)
	CreateSession(pin string, nonce string) (string, error)
	SetPINCode(pinCode string)
	ForgetSession(sessionID string)
	SessionIsValid(sessionID string) bool
	Lock()
}
