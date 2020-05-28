package api

import (
	"encoding/json"

	"github.com/nordcloud/termination-detector/internal/process"

	"github.com/pkg/errors"
)

type Process struct {
	ID           string        `json:"id"`
	State        process.State `json:"state"`
	StateMessage *string       `json:"stateMessage,omitempty"`
}

func (process Process) JSON() string {
	marshalledProcess, err := json.Marshal(process)
	if err != nil {
		panic(errors.Wrapf(err, "failed to marshal task: %+v", marshalledProcess))
	}
	return string(marshalledProcess)
}

func ConvertInternalProcess(proc process.Process) Process {
	return Process{
		ID:           proc.ID,
		State:        proc.State,
		StateMessage: proc.StateMessage,
	}
}
