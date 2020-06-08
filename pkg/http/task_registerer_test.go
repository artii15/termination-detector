package http_test

import (
	"errors"
	"net/http"
	"testing"
	"time"

	internalHTTP "github.com/artii15/termination-detector/pkg/http"
	"github.com/artii15/termination-detector/pkg/task"
	"github.com/stretchr/testify/assert"
)

type taskRegistererWithMocks struct {
	requestExecutor *requestExecutorMock
	taskRegisterer  *internalHTTP.TaskRegisterer
}

func newTaskRegistererWithMocks() *taskRegistererWithMocks {
	requestExecutor := new(requestExecutorMock)
	taskRegisterer := internalHTTP.NewTaskRegisterer(requestExecutor)
	return &taskRegistererWithMocks{
		requestExecutor: requestExecutor,
		taskRegisterer:  taskRegisterer,
	}
}

func TestTaskRegisterer_Register(t *testing.T) {
	taskRegistererAndMocks := newTaskRegistererWithMocks()
	taskExpirationTime := time.Now().Add(time.Hour)
	taskToRegister := internalHTTP.Task{ExpirationTime: taskExpirationTime}
	taskRegistrationData := task.RegistrationData{
		ID: task.ID{
			ProcessID: "1",
			TaskID:    "2",
		},
		ExpirationTime: taskExpirationTime,
	}
	taskRegistererAndMocks.requestExecutor.On("ExecuteRequest", internalHTTP.Request{
		Method:       internalHTTP.MethodPut,
		ResourcePath: internalHTTP.ResourcePathTask,
		Body:         taskToRegister.JSON(),
		PathParameters: map[internalHTTP.PathParameter]string{
			internalHTTP.PathParameterProcessID: taskRegistrationData.ID.ProcessID,
			internalHTTP.PathParameterTaskID:    taskRegistrationData.ID.TaskID,
		},
	}).Return(internalHTTP.Response{
		StatusCode: http.StatusCreated,
	}, nil)

	registrationStatus, err := taskRegistererAndMocks.taskRegisterer.Register(taskRegistrationData)
	assert.NoError(t, err)
	assert.Equal(t, task.RegistrationResultCreated, registrationStatus)
}

func TestTaskRegisterer_Register_TaskAlreadyRegistered(t *testing.T) {
	taskRegistererAndMocks := newTaskRegistererWithMocks()
	taskExpirationTime := time.Now().Add(time.Hour)
	taskToRegister := internalHTTP.Task{ExpirationTime: taskExpirationTime}
	taskRegistrationData := task.RegistrationData{
		ID: task.ID{
			ProcessID: "1",
			TaskID:    "2",
		},
		ExpirationTime: taskExpirationTime,
	}
	taskRegistererAndMocks.requestExecutor.On("ExecuteRequest", internalHTTP.Request{
		Method:       internalHTTP.MethodPut,
		ResourcePath: internalHTTP.ResourcePathTask,
		Body:         taskToRegister.JSON(),
		PathParameters: map[internalHTTP.PathParameter]string{
			internalHTTP.PathParameterProcessID: taskRegistrationData.ID.ProcessID,
			internalHTTP.PathParameterTaskID:    taskRegistrationData.ID.TaskID,
		},
	}).Return(internalHTTP.Response{
		StatusCode: http.StatusConflict,
	}, nil)

	registrationStatus, err := taskRegistererAndMocks.taskRegisterer.Register(taskRegistrationData)
	assert.NoError(t, err)
	assert.Equal(t, task.RegistrationResultAlreadyRegistered, registrationStatus)
}

func TestTaskRegisterer_Register_UnexpectedResponseStatus(t *testing.T) {
	taskRegistererAndMocks := newTaskRegistererWithMocks()
	taskExpirationTime := time.Now().Add(time.Hour)
	taskToRegister := internalHTTP.Task{ExpirationTime: taskExpirationTime}
	taskRegistrationData := task.RegistrationData{
		ID: task.ID{
			ProcessID: "1",
			TaskID:    "2",
		},
		ExpirationTime: taskExpirationTime,
	}
	taskRegistererAndMocks.requestExecutor.On("ExecuteRequest", internalHTTP.Request{
		Method:       internalHTTP.MethodPut,
		ResourcePath: internalHTTP.ResourcePathTask,
		Body:         taskToRegister.JSON(),
		PathParameters: map[internalHTTP.PathParameter]string{
			internalHTTP.PathParameterProcessID: taskRegistrationData.ID.ProcessID,
			internalHTTP.PathParameterTaskID:    taskRegistrationData.ID.TaskID,
		},
	}).Return(internalHTTP.Response{
		StatusCode: http.StatusInternalServerError,
	}, nil)

	_, err := taskRegistererAndMocks.taskRegisterer.Register(taskRegistrationData)
	assert.Error(t, err)
}

func TestTaskRegisterer_Register_ExecutorError(t *testing.T) {
	taskRegistererAndMocks := newTaskRegistererWithMocks()
	taskExpirationTime := time.Now().Add(time.Hour)
	taskToRegister := internalHTTP.Task{ExpirationTime: taskExpirationTime}
	taskRegistrationData := task.RegistrationData{
		ID: task.ID{
			ProcessID: "1",
			TaskID:    "2",
		},
		ExpirationTime: taskExpirationTime,
	}
	taskRegistererAndMocks.requestExecutor.On("ExecuteRequest", internalHTTP.Request{
		Method:       internalHTTP.MethodPut,
		ResourcePath: internalHTTP.ResourcePathTask,
		Body:         taskToRegister.JSON(),
		PathParameters: map[internalHTTP.PathParameter]string{
			internalHTTP.PathParameterProcessID: taskRegistrationData.ID.ProcessID,
			internalHTTP.PathParameterTaskID:    taskRegistrationData.ID.TaskID,
		},
	}).Return(internalHTTP.Response{}, errors.New("error"))

	_, err := taskRegistererAndMocks.taskRegisterer.Register(taskRegistrationData)
	assert.Error(t, err)
}
