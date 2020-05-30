package task

import (
	"time"
)

type RegistrationResult string

const (
	RegistrationResultCreated           RegistrationResult = "CREATED"
	RegistrationResultAlreadyRegistered RegistrationResult = "ALREADY_REGISTERED"
)

type RegistrationData struct {
	ID             ID
	ExpirationTime time.Time
}

type Registerer interface {
	Register(registrationData RegistrationData) (RegistrationResult, error)
}
