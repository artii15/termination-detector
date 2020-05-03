package api

import (
	"encoding/json"
	"time"

	"github.com/pkg/errors"
)

const (
	TaskIDPathParameter = "task_id"
)

type Task struct {
	ExpirationTime time.Time `json:"expirationTime"`
}

func (task *Task) JSON() string {
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
