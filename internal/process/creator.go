package process

type CreationStatus string

const (
	CreationStatusNew        CreationStatus = "NEW"
	CreationStatusOverridden CreationStatus = "OVERRIDDEN"
	CreationStatusConflict   CreationStatus = "CONFLICT"
)

type Creator interface {
	Create(process Process) (CreationStatus, error)
}
