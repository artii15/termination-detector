package process

type State string

const (
	StateCompleted = "COMPLETED"
	StateCreated   = "CREATED"
	StateError     = "ERROR"
)

type Process struct {
	ID           string
	State        State
	StateMessage *string
}
