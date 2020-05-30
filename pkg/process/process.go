package process

type State string

const (
	StateCompleted State = "COMPLETED"
	StateCreated   State = "CREATED"
	StateError     State = "ERROR"

	TimedOutErrorMessage = "process timed out"
)

type Process struct {
	ID           string
	State        State
	StateMessage *string
}
