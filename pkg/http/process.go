package http

import (
	"encoding/json"

	"github.com/artii15/termination-detector/pkg/process"
	"github.com/pkg/errors"
)

type Process struct {
	ID           string        `json:"id"`
	State        process.State `json:"state"`
	StateMessage *string       `json:"stateMessage,omitempty"`
}

func (proc Process) JSON() string {
	marshalledProcess, err := json.Marshal(proc)
	if err != nil {
		panic(errors.Wrapf(err, "failed to marshal task: %+v", marshalledProcess))
	}
	return string(marshalledProcess)
}

func (proc *Process) optionalInternalProcess() *process.Process {
	if proc == nil {
		return nil
	}
	converted := proc.internalProcess()
	return &converted
}

func (proc Process) internalProcess() process.Process {
	return process.Process{
		ID:           proc.ID,
		State:        proc.State,
		StateMessage: proc.StateMessage,
	}
}

func ConvertInternalToHTTPProcess(proc process.Process) Process {
	return Process{
		ID:           proc.ID,
		State:        proc.State,
		StateMessage: proc.StateMessage,
	}
}
