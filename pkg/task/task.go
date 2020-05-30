package task

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
