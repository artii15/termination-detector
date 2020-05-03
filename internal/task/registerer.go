package task

import "time"

type RegistrationResult string

const (
	RegistrationResultCreated            RegistrationResult = "CREATED"
	RegistrationResultNotChanged         RegistrationResult = "NOT_CHANGED"
	RegistrationResultChanged            RegistrationResult = "CHANGED"
	RegistrationResultErrorTerminalState RegistrationResult = "ALREADY_IN_TERMINAL_STATE"
)

type RegistrationData struct {
	ID             ID
	ExpirationTime time.Time
}

type Registerer interface {
	Register(registrationData RegistrationData) (RegistrationResult, error)
}
