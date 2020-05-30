package task

type CompleteRequest struct {
	ID
	State   State
	Message *string
}

type CompletingResult string

const (
	CompletingResultConflict  CompletingResult = "CONFLICT"
	CompletingResultCompleted CompletingResult = "COMPLETED"
)

type Completer interface {
	Complete(request CompleteRequest) (CompletingResult, error)
}
