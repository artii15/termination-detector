package api

import (
	"encoding/json"
	"time"

	"github.com/pkg/errors"
)

const (
	ProcessIDPathParameter = "process_id"
	TaskIDPathParameter    = "task_id"
)

type Task struct {
	ExpirationTime time.Time `json:"expirationTime"`
}

func (task Task) JSON() string {
	marshalledTask, err := json.Marshal(task)
	if err != nil {
		panic(errors.Wrapf(err, "failed to marshal task: %+v", marshalledTask))
	}
	return string(marshalledTask)
}

func UnmarshalTask(marshalledTask string) (task Task, err error) {
	err = json.Unmarshal([]byte(marshalledTask), &task)
	return
}

type CompletionState string

const (
	CompletionStateError     CompletionState = "ERROR"
	CompletionStateCompleted CompletionState = "COMPLETED"
)

type Completion struct {
	State        CompletionState `json:"state"`
	ErrorMessage *string         `json:"errorMessage,omitempty"`
}

func (completion Completion) JSON() string {
	marshalled, err := json.Marshal(completion)
	if err != nil {
		panic(errors.Wrapf(err, "failed to marshal task completion: %+v", completion))
	}
	return string(marshalled)
}

func UnmarshalCompletion(marshalledCompletion string) (completion Completion, err error) {
	err = json.Unmarshal([]byte(marshalledCompletion), &completion)
	return
}
