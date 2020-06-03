package http

import (
	"fmt"
	"net/http"

	"github.com/nordcloud/termination-detector/pkg/task"
)

type TaskRegisterer struct {
	requestExecutor requestExecutor
}

func NewTaskRegisterer(requestExecutor requestExecutor) *TaskRegisterer {
	return &TaskRegisterer{
		requestExecutor: requestExecutor,
	}
}

func (registerer *TaskRegisterer) Register(registrationData task.RegistrationData) (task.RegistrationResult, error) {
	taskToRegister := Task{
		ExpirationTime: registrationData.ExpirationTime,
	}
	response, err := registerer.requestExecutor.ExecuteRequest(Request{
		Method:       MethodPut,
		ResourcePath: ResourcePathTask,
		Body:         taskToRegister.JSON(),
		PathParameters: map[PathParameter]string{
			PathParameterProcessID: registrationData.ID.ProcessID,
			PathParameterTaskID:    registrationData.ID.TaskID,
		},
	})
	if err != nil {
		return "", err
	}
	switch response.StatusCode {
	case http.StatusCreated:
		return task.RegistrationResultCreated, nil
	case http.StatusConflict:
		return task.RegistrationResultAlreadyRegistered, nil
	default:
		return "", fmt.Errorf("unknown task registration result: %d %s", response.StatusCode, response.Body)
	}
}
