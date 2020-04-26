package task

import "time"

type State string

const (
	StateRegistered State = "REGISTERED"
)

type Task struct {
	RegistrationData
	State State
}

type RegistrationData struct {
	ID             string
	ProcessID      string
	ExpirationTime *time.Time
	IsLastTask     bool
}
