package process

type State string

const (
	StateCreated  State = "CREATED"
	StateFinished State = "FINISHED"
	StateError    State = "ERROR"
)

type Process struct {
	ID               string
	State            State
	StateDescription *string
}
