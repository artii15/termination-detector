package handlers_test

import (
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/artii15/termination-detector/internal/api/handlers"
	internalHTTP "github.com/artii15/termination-detector/pkg/http"
	"github.com/artii15/termination-detector/pkg/task"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type taskRegistererMock struct {
	mock.Mock
}

func (registerer *taskRegistererMock) Register(registrationData task.RegistrationData) (task.RegistrationResult, error) {
	args := registerer.Called(registrationData)
	return args.Get(0).(task.RegistrationResult), args.Error(1)
}

type putTaskReqHandlerWithMocks struct {
	request            internalHTTP.Request
	registrationData   task.RegistrationData
	taskRegistererMock *taskRegistererMock
	handler            *handlers.PutTaskRequestHandler
}

func (handlerAndMocks *putTaskReqHandlerWithMocks) assertExpectations(t *testing.T) {
	handlerAndMocks.taskRegistererMock.AssertExpectations(t)
}

func newPutTaskReqHandlerWithMocks() *putTaskReqHandlerWithMocks {
	taskRegisterer := new(taskRegistererMock)
	apiTask := internalHTTP.Task{
		ExpirationTime: time.Now().Add(time.Hour).UTC(),
	}
	taskID := "1"
	processID := "2"
	return &putTaskReqHandlerWithMocks{
		request: internalHTTP.Request{
			PathParameters: map[internalHTTP.PathParameter]string{
				internalHTTP.PathParameterTaskID:    taskID,
				internalHTTP.PathParameterProcessID: processID,
			},
			Body: apiTask.JSON(),
		},
		registrationData: task.RegistrationData{
			ID: task.ID{
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
		Return(task.RegistrationResultCreated, nil)

	response, err := handlerAndMocks.handler.HandleRequest(handlerAndMocks.request)
	assert.NoError(t, err)
	handlerAndMocks.assertExpectations(t)
	assert.Equal(t, internalHTTP.Response{
		StatusCode: http.StatusCreated,
		Headers: map[string]string{
			internalHTTP.ContentTypeHeaderName: internalHTTP.ContentTypeApplicationJSON,
		},
		Body: handlerAndMocks.request.Body,
	}, response)
}

func TestPutTaskRequestHandler_HandleRequest_DuplicatedLastTask(t *testing.T) {
	handlerAndMocks := newPutTaskReqHandlerWithMocks()

	handlerAndMocks.taskRegistererMock.On("Register", handlerAndMocks.registrationData).
		Return(task.RegistrationResultAlreadyRegistered, nil)

	response, err := handlerAndMocks.handler.HandleRequest(handlerAndMocks.request)
	assert.NoError(t, err)
	handlerAndMocks.assertExpectations(t)
	assert.Equal(t, internalHTTP.Response{
		StatusCode: http.StatusConflict,
		Headers: map[string]string{
			internalHTTP.ContentTypeHeaderName: internalHTTP.ContentTypeTextPlain,
		},
		Body: handlers.TaskAlreadyCreatedErrorMessage,
	}, response)
}

func TestPutTaskRequestHandler_HandleRequest_UnknownRegistrationResult(t *testing.T) {
	handlerAndMocks := newPutTaskReqHandlerWithMocks()

	handlerAndMocks.taskRegistererMock.On("Register", handlerAndMocks.registrationData).
		Return(task.RegistrationResult("unknown"), nil)

	_, err := handlerAndMocks.handler.HandleRequest(handlerAndMocks.request)
	handlerAndMocks.assertExpectations(t)
	assert.Error(t, err)
}

func TestPutTaskRequestHandler_HandleRequest_RegistrationFailure(t *testing.T) {
	handlerAndMocks := newPutTaskReqHandlerWithMocks()

	handlerAndMocks.taskRegistererMock.On("Register", handlerAndMocks.registrationData).
		Return(task.RegistrationResult(""), errors.New("error"))

	_, err := handlerAndMocks.handler.HandleRequest(handlerAndMocks.request)
	handlerAndMocks.assertExpectations(t)
	assert.Error(t, err)
}

func TestPutTaskRequestHandler_HandleRequest_InvalidBody(t *testing.T) {
	handlerAndMocks := newPutTaskReqHandlerWithMocks()
	handlerAndMocks.request.Body = ""

	response, err := handlerAndMocks.handler.HandleRequest(handlerAndMocks.request)
	assert.NoError(t, err)
	assert.Equal(t, internalHTTP.Response{
		StatusCode: http.StatusBadRequest,
		Body:       handlers.InvalidPayloadErrorMessage,
		Headers: map[string]string{
			internalHTTP.ContentTypeHeaderName: internalHTTP.ContentTypeTextPlain,
		},
	}, response)
}
