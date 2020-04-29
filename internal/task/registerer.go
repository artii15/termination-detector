package task

type RegistrationResult string

const (
	RegistrationResultCreated            RegistrationResult = "CREATED"
	RegistrationResultNotChanged         RegistrationResult = "NOT_CHANGED"
	RegistrationResultChanged            RegistrationResult = "CHANGED"
	RegistrationResultErrorTerminalState RegistrationResult = "ALREADY_IN_TERMINAL_STATE"
)

type Registerer interface {
	Register(registrationData RegistrationData) (RegistrationResult, error)
}
