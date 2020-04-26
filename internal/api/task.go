package api

import (
	"encoding/json"
	"time"
)

const (
	TaskIDPathParameter = "task-id"
)

type Task struct {
	ExpirationTime *time.Time `json:"expirationTime"`
	IsLastTask     bool       `json:"isLastTask"`
}

func UnmarshalTask(marshalledTask string) (task Task, err error) {
	err = json.Unmarshal([]byte(marshalledTask), &task)
	return
}
