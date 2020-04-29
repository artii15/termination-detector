package task

import "time"

type State string

const (
	StateAborted  State = "ABORTED"
	StateCreated  State = "CREATED"
	StateFinished State = "FINISHED"
)

type Task struct {
	RegistrationData
	State State
}

type RegistrationData struct {
	ID             string
	ProcessID      string
	ExpirationTime time.Time
}

func (data RegistrationData) Equals(other RegistrationData) bool {
	return data.ID == other.ID && data.ProcessID == other.ProcessID && data.ExpirationTime == other.ExpirationTime
}
