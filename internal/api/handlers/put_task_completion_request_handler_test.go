package handlers_test

import (
	"errors"
	"net/http"
	"testing"

	task2 "github.com/nordcloud/termination-detector/pkg/task"

	"github.com/stretchr/testify/assert"

	"github.com/stretchr/testify/mock"

	"github.com/nordcloud/termination-detector/internal/api"
	"github.com/nordcloud/termination-detector/internal/api/handlers"
)

type taskCompleterMock struct {
	mock.Mock
}

func (completer *taskCompleterMock) Complete(request task2.CompleteRequest) (task2.CompletingResult, error) {
	args := completer.Called(request)
	return args.Get(0).(task2.CompletingResult), args.Error(1)
}

type putTaskCompletionReqHandlerWithMocks struct {
	request       api.Request
	completion    api.Completion
	completerMock *taskCompleterMock
	taskID        task2.ID
	handler       *handlers.PutTaskCompletionRequestHandler
}

func (handlerAndMocks *putTaskCompletionReqHandlerWithMocks) assertExpectations(t *testing.T) {
	handlerAndMocks.completerMock.AssertExpectations(t)
}

func newPutTaskCompletionReqHandlerWithMocks(completion api.Completion) *putTaskCompletionReqHandlerWithMocks {
	taskID := task2.ID{
		ProcessID: "2",
		TaskID:    "1",
	}
	completerMock := new(taskCompleterMock)
	return &putTaskCompletionReqHandlerWithMocks{
		request: api.Request{
			PathParameters: map[string]string{
				api.TaskIDPathParameter:    taskID.TaskID,
				api.ProcessIDPathParameter: taskID.ProcessID,
			},
			Body: completion.JSON(),
		},
		completion:    completion,
		handler:       handlers.NewPutTaskCompletionRequestHandler(completerMock),
		completerMock: completerMock,
		taskID:        taskID,
	}
}

func TestPutTaskCompletionRequestHandler_HandleRequest(t *testing.T) {
	completion := api.Completion{State: api.CompletionStateCompleted}
	handlerAndMocks := newPutTaskCompletionReqHandlerWithMocks(completion)
	handlerAndMocks.completerMock.On("Complete", task2.CompleteRequest{
		ID:    handlerAndMocks.taskID,
		State: task2.StateFinished,
	}).Return(task2.CompletingResultCompleted, nil)

	response, err := handlerAndMocks.handler.HandleRequest(handlerAndMocks.request)
	handlerAndMocks.assertExpectations(t)
	assert.NoError(t, err)
	assert.Equal(t, api.Response{
		StatusCode: http.StatusCreated,
		Body:       handlerAndMocks.request.Body,
		Headers:    map[string]string{api.ContentTypeHeaderName: api.ContentTypeApplicationJSON},
	}, response)
}

func TestPutTaskCompletionRequestHandler_HandleRequest_InvalidPayload(t *testing.T) {
	handler := handlers.NewPutTaskCompletionRequestHandler(new(taskCompleterMock))
	response, err := handler.HandleRequest(api.Request{
		Body: "",
	})

	assert.NoError(t, err)
	assert.Equal(t, api.Response{
		StatusCode: http.StatusBadRequest,
		Body:       handlers.InvalidPayloadErrorMessage,
		Headers:    map[string]string{api.ContentTypeHeaderName: api.ContentTypeTextPlain},
	}, response)
}

func TestPutTaskCompletionRequestHandler_HandleRequest_UnknownCompletionState(t *testing.T) {
	handler := handlers.NewPutTaskCompletionRequestHandler(new(taskCompleterMock))
	completion := api.Completion{State: api.CompletionState("invalid")}
	response, err := handler.HandleRequest(api.Request{
		Body: completion.JSON(),
	})

	assert.NoError(t, err)
	assert.Equal(t, api.Response{
		StatusCode: http.StatusBadRequest,
		Body:       handlers.UnknownCompletionStateMsg,
		Headers:    map[string]string{api.ContentTypeHeaderName: api.ContentTypeTextPlain},
	}, response)
}

func TestPutTaskCompletionRequestHandler_HandleRequest_CompleteWithError(t *testing.T) {
	errorMsg := "error"
	completion := api.Completion{State: api.CompletionStateError, ErrorMessage: &errorMsg}
	handlerAndMocks := newPutTaskCompletionReqHandlerWithMocks(completion)
	handlerAndMocks.completerMock.On("Complete", task2.CompleteRequest{
		ID:      handlerAndMocks.taskID,
		State:   task2.StateAborted,
		Message: &errorMsg,
	}).Return(task2.CompletingResultCompleted, nil)

	response, err := handlerAndMocks.handler.HandleRequest(handlerAndMocks.request)
	handlerAndMocks.assertExpectations(t)
	assert.NoError(t, err)
	assert.Equal(t, api.Response{
		StatusCode: http.StatusCreated,
		Body:       handlerAndMocks.request.Body,
		Headers:    map[string]string{api.ContentTypeHeaderName: api.ContentTypeApplicationJSON},
	}, response)
}

func TestPutTaskCompletionRequestHandler_HandleRequest_TaskStateConflict(t *testing.T) {
	completion := api.Completion{State: api.CompletionStateCompleted}
	handlerAndMocks := newPutTaskCompletionReqHandlerWithMocks(completion)
	handlerAndMocks.completerMock.On("Complete", task2.CompleteRequest{
		ID:    handlerAndMocks.taskID,
		State: task2.StateFinished,
	}).Return(task2.CompletingResultConflict, nil)

	response, err := handlerAndMocks.handler.HandleRequest(handlerAndMocks.request)
	handlerAndMocks.assertExpectations(t)
	assert.NoError(t, err)
	assert.Equal(t, api.Response{
		StatusCode: http.StatusConflict,
		Body:       handlers.ConflictingTaskCompletionMsg,
		Headers:    map[string]string{api.ContentTypeHeaderName: api.ContentTypeTextPlain},
	}, response)
}

func TestPutTaskCompletionRequestHandler_HandleRequest_UnknownCompletionResult(t *testing.T) {
	completion := api.Completion{State: api.CompletionStateCompleted}
	handlerAndMocks := newPutTaskCompletionReqHandlerWithMocks(completion)
	handlerAndMocks.completerMock.On("Complete", task2.CompleteRequest{
		ID:    handlerAndMocks.taskID,
		State: task2.StateFinished,
	}).Return(task2.CompletingResult("unknown"), nil)

	response, err := handlerAndMocks.handler.HandleRequest(handlerAndMocks.request)
	handlerAndMocks.assertExpectations(t)
	assert.NoError(t, err)
	assert.Equal(t, api.Response{
		StatusCode: http.StatusInternalServerError,
		Body:       handlers.UnknownErrorMsg,
		Headers:    map[string]string{api.ContentTypeHeaderName: api.ContentTypeTextPlain},
	}, response)
}

func TestPutTaskCompletionRequestHandler_HandleRequest_CompletionError(t *testing.T) {
	completion := api.Completion{State: api.CompletionStateCompleted}
	handlerAndMocks := newPutTaskCompletionReqHandlerWithMocks(completion)
	handlerAndMocks.completerMock.On("Complete", task2.CompleteRequest{
		ID:    handlerAndMocks.taskID,
		State: task2.StateFinished,
	}).Return(task2.CompletingResultCompleted, errors.New("error"))

	_, err := handlerAndMocks.handler.HandleRequest(handlerAndMocks.request)
	handlerAndMocks.assertExpectations(t)
	assert.Error(t, err)
}
