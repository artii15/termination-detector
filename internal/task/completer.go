package task

type Completion struct {
	State   State
	Message *string
}

type CompleteRequest struct {
	ProcessID string
	TaskID    string
	Completion
}

type Completer interface {
	Complete(request CompleteRequest) error
}
