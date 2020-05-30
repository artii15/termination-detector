package task

import "time"

type State string

const (
	StateAborted  State = "ABORTED"
	StateCreated  State = "CREATED"
	StateFinished State = "FINISHED"
)

type ID struct {
	ProcessID string
	TaskID    string
}

type Task struct {
	ID
	State          State
	ExpirationTime time.Time
	StateMessage   *string
}
