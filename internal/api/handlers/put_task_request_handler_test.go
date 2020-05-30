package handlers_test

import (
	"errors"
	"net/http"
	"testing"
	"time"

	task2 "github.com/nordcloud/termination-detector/pkg/task"

	"github.com/nordcloud/termination-detector/internal/api"
	"github.com/nordcloud/termination-detector/internal/api/handlers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type taskRegistererMock struct {
	mock.Mock
}

func (registerer *taskRegistererMock) Register(registrationData task2.RegistrationData) (task2.RegistrationResult, error) {
	args := registerer.Called(registrationData)
	return args.Get(0).(task2.RegistrationResult), args.Error(1)
}

type putTaskReqHandlerWithMocks struct {
	request            api.Request
	registrationData   task2.RegistrationData
	taskRegistererMock *taskRegistererMock
	handler            *handlers.PutTaskRequestHandler
}

func (handlerAndMocks *putTaskReqHandlerWithMocks) assertExpectations(t *testing.T) {
	handlerAndMocks.taskRegistererMock.AssertExpectations(t)
}

func newPutTaskReqHandlerWithMocks() *putTaskReqHandlerWithMocks {
	taskRegisterer := new(taskRegistererMock)
	apiTask := api.Task{
		ExpirationTime: time.Now().Add(time.Hour).UTC(),
	}
	taskID := "1"
	processID := "2"
	return &putTaskReqHandlerWithMocks{
		request: api.Request{
			PathParameters: map[string]string{
				api.TaskIDPathParameter:    taskID,
				api.ProcessIDPathParameter: processID,
			},
			Body: apiTask.JSON(),
		},
		registrationData: task2.RegistrationData{
			ID: task2.ID{
				ProcessID: processID,
				TaskID:    taskID,
			},
			ExpirationTime: apiTask.ExpirationTime,
		},
		taskRegistererMock: taskRegisterer,
		handler:            handlers.NewPutTaskRequestHandler(taskRegisterer),
	}
}

func TestPutTaskRequestHandler_HandleRequest_TaskCreated(t *testing.T) {
	handlerAndMocks := newPutTaskReqHandlerWithMocks()

	handlerAndMocks.taskRegistererMock.On("Register", handlerAndMocks.registrationData).
		Return(task2.RegistrationResultCreated, nil)

	response, err := handlerAndMocks.handler.HandleRequest(handlerAndMocks.request)
	assert.NoError(t, err)
	handlerAndMocks.assertExpectations(t)
	assert.Equal(t, api.Response{
		StatusCode: http.StatusCreated,
		Headers: map[string]string{
			api.ContentTypeHeaderName: api.ContentTypeApplicationJSON,
		},
		Body: handlerAndMocks.request.Body,
	}, response)
}

func TestPutTaskRequestHandler_HandleRequest_DuplicatedLastTask(t *testing.T) {
	handlerAndMocks := newPutTaskReqHandlerWithMocks()

	handlerAndMocks.taskRegistererMock.On("Register", handlerAndMocks.registrationData).
		Return(task2.RegistrationResultAlreadyRegistered, nil)

	response, err := handlerAndMocks.handler.HandleRequest(handlerAndMocks.request)
	assert.NoError(t, err)
	handlerAndMocks.assertExpectations(t)
	assert.Equal(t, api.Response{
		StatusCode: http.StatusConflict,
		Headers: map[string]string{
			api.ContentTypeHeaderName: api.ContentTypeTextPlain,
		},
		Body: handlers.TaskAlreadyCreatedErrorMessage,
	}, response)
}

func TestPutTaskRequestHandler_HandleRequest_UnknownRegistrationResult(t *testing.T) {
	handlerAndMocks := newPutTaskReqHandlerWithMocks()

	handlerAndMocks.taskRegistererMock.On("Register", handlerAndMocks.registrationData).
		Return(task2.RegistrationResult("unknown"), nil)

	_, err := handlerAndMocks.handler.HandleRequest(handlerAndMocks.request)
	handlerAndMocks.assertExpectations(t)
	assert.Error(t, err)
}

func TestPutTaskRequestHandler_HandleRequest_RegistrationFailure(t *testing.T) {
	handlerAndMocks := newPutTaskReqHandlerWithMocks()

	handlerAndMocks.taskRegistererMock.On("Register", handlerAndMocks.registrationData).
		Return(task2.RegistrationResult(""), errors.New("error"))

	_, err := handlerAndMocks.handler.HandleRequest(handlerAndMocks.request)
	handlerAndMocks.assertExpectations(t)
	assert.Error(t, err)
}

func TestPutTaskRequestHandler_HandleRequest_InvalidBody(t *testing.T) {
	handlerAndMocks := newPutTaskReqHandlerWithMocks()
	handlerAndMocks.request.Body = ""

	response, err := handlerAndMocks.handler.HandleRequest(handlerAndMocks.request)
	assert.NoError(t, err)
	assert.Equal(t, api.Response{
		StatusCode: http.StatusBadRequest,
		Body:       handlers.InvalidPayloadErrorMessage,
		Headers: map[string]string{
			api.ContentTypeHeaderName: api.ContentTypeTextPlain,
		},
	}, response)
}
