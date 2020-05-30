package process

type Getter interface {
	Get(processID string) (*Process, error)
}
