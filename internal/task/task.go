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

func (id ID) Equals(other ID) bool {
	return id.ProcessID == other.ProcessID && id.TaskID == other.TaskID
}

type Task struct {
	ID
	State          State
	ExpirationTime time.Time
}
