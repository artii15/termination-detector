package task

type RegistrationResult string

const (
	RegistrationResultCreated           RegistrationResult = "REGISTERED"
	RegistrationResultNotChanged        RegistrationResult = "NOT_CHANGED"
	RegistrationResultChanged           RegistrationResult = "CHANGED"
	RegistrationResultDuplicateLastTask RegistrationResult = "DUPLICATE_LAST_TASK"
)

type Registerer interface {
	Register(registrationData RegistrationData) (RegistrationResult, error)
}
