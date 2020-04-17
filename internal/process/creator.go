package process

type CreationResult struct {
	AlreadyExistsInConflictingState bool
}

type Creator interface {
	Create(process Process) (CreationResult, error)
}
