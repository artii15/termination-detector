package handlers_test

import (
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/nordcloud/termination-detector/internal/api"
	"github.com/nordcloud/termination-detector/internal/api/handlers"
	"github.com/nordcloud/termination-detector/internal/task"
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
	request            events.APIGatewayProxyRequest
	registrationData   task.RegistrationData
	taskRegistererMock *taskRegistererMock
	handler            *handlers.PutTaskRequestHandler
}

func newPutTaskReqHandlerWithMocks() *putTaskReqHandlerWithMocks {
	taskRegisterer := new(taskRegistererMock)
	apiTask := api.Task{
		ExpirationTime: time.Now().Add(time.Hour).UTC(),
	}
	taskID := "1"
	processID := "2"
	return &putTaskReqHandlerWithMocks{
		request: events.APIGatewayProxyRequest{
			PathParameters: map[string]string{
				api.TaskIDPathParameter:    taskID,
				api.ProcessIDPathParameter: processID,
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

func TestPutTaskRequestHandler_HandleRequest_TaskChanged(t *testing.T) {
	handlerAndMocks := newPutTaskReqHandlerWithMocks()

	handlerAndMocks.taskRegistererMock.On("Register", handlerAndMocks.registrationData).
		Return(task.RegistrationResultChanged, nil)

	response, err := handlerAndMocks.handler.HandleRequest(handlerAndMocks.request)
	assert.NoError(t, err)
	assert.Equal(t, events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Headers: map[string]string{
			api.ContentTypeHeaderName: api.ContentTypeApplicationJSON,
		},
		Body: handlerAndMocks.request.Body,
	}, response)
}

func TestPutTaskRequestHandler_HandleRequest_TaskCreated(t *testing.T) {
	handlerAndMocks := newPutTaskReqHandlerWithMocks()

	handlerAndMocks.taskRegistererMock.On("Register", handlerAndMocks.registrationData).
		Return(task.RegistrationResultCreated, nil)

	response, err := handlerAndMocks.handler.HandleRequest(handlerAndMocks.request)
	assert.NoError(t, err)
	assert.Equal(t, events.APIGatewayProxyResponse{
		StatusCode: http.StatusCreated,
		Headers: map[string]string{
			api.ContentTypeHeaderName: api.ContentTypeApplicationJSON,
		},
		Body: handlerAndMocks.request.Body,
	}, response)
}

func TestPutTaskRequestHandler_HandleRequest_TaskNotChanged(t *testing.T) {
	handlerAndMocks := newPutTaskReqHandlerWithMocks()

	handlerAndMocks.taskRegistererMock.On("Register", handlerAndMocks.registrationData).
		Return(task.RegistrationResultNotChanged, nil)

	response, err := handlerAndMocks.handler.HandleRequest(handlerAndMocks.request)
	assert.NoError(t, err)
	assert.Equal(t, events.APIGatewayProxyResponse{
		StatusCode: http.StatusNoContent,
	}, response)
}

func TestPutTaskRequestHandler_HandleRequest_DuplicatedLastTask(t *testing.T) {
	handlerAndMocks := newPutTaskReqHandlerWithMocks()

	handlerAndMocks.taskRegistererMock.On("Register", handlerAndMocks.registrationData).
		Return(task.RegistrationResultErrorTerminalState, nil)

	response, err := handlerAndMocks.handler.HandleRequest(handlerAndMocks.request)
	assert.NoError(t, err)
	assert.Equal(t, events.APIGatewayProxyResponse{
		StatusCode: http.StatusConflict,
		Headers: map[string]string{
			api.ContentTypeHeaderName: api.ContentTypeTextPlain,
		},
		Body: handlers.TaskInTerminalStateErrorMessage,
	}, response)
}

func TestPutTaskRequestHandler_HandleRequest_UnknownRegistrationResult(t *testing.T) {
	handlerAndMocks := newPutTaskReqHandlerWithMocks()

	handlerAndMocks.taskRegistererMock.On("Register", handlerAndMocks.registrationData).
		Return(task.RegistrationResult("unknown"), nil)

	_, err := handlerAndMocks.handler.HandleRequest(handlerAndMocks.request)
	assert.Error(t, err)
}

func TestPutTaskRequestHandler_HandleRequest_RegistrationFailure(t *testing.T) {
	handlerAndMocks := newPutTaskReqHandlerWithMocks()

	handlerAndMocks.taskRegistererMock.On("Register", handlerAndMocks.registrationData).
		Return(task.RegistrationResult(""), errors.New("error"))

	_, err := handlerAndMocks.handler.HandleRequest(handlerAndMocks.request)
	assert.Error(t, err)
}
