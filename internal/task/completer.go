package task

type CompleteRequest struct {
	ID
	State   State
	Message *string
}

type Completer interface {
	Complete(request CompleteRequest) error
}
