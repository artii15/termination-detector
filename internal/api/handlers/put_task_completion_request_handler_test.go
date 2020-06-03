package handlers_test

import (
	"errors"
	"net/http"
	"testing"

	"github.com/nordcloud/termination-detector/internal/api"
	"github.com/nordcloud/termination-detector/internal/api/handlers"
	internalHTTP "github.com/nordcloud/termination-detector/pkg/http"
	"github.com/nordcloud/termination-detector/pkg/task"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type taskCompleterMock struct {
	mock.Mock
}

func (completer *taskCompleterMock) Complete(request task.CompleteRequest) (task.CompletingResult, error) {
	args := completer.Called(request)
	return args.Get(0).(task.CompletingResult), args.Error(1)
}

type putTaskCompletionReqHandlerWithMocks struct {
	request       internalHTTP.Request
	completion    internalHTTP.Completion
	completerMock *taskCompleterMock
	taskID        task.ID
	handler       *handlers.PutTaskCompletionRequestHandler
}

func (handlerAndMocks *putTaskCompletionReqHandlerWithMocks) assertExpectations(t *testing.T) {
	handlerAndMocks.completerMock.AssertExpectations(t)
}

func newPutTaskCompletionReqHandlerWithMocks(completion internalHTTP.Completion) *putTaskCompletionReqHandlerWithMocks {
	taskID := task.ID{
		ProcessID: "2",
		TaskID:    "1",
	}
	completerMock := new(taskCompleterMock)
	return &putTaskCompletionReqHandlerWithMocks{
		request: internalHTTP.Request{
			PathParameters: map[internalHTTP.PathParameter]string{
				internalHTTP.PathParameterTaskID:    taskID.TaskID,
				internalHTTP.PathParameterProcessID: taskID.ProcessID,
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
	completion := internalHTTP.Completion{State: internalHTTP.CompletionStateCompleted}
	handlerAndMocks := newPutTaskCompletionReqHandlerWithMocks(completion)
	handlerAndMocks.completerMock.On("Complete", task.CompleteRequest{
		ID:    handlerAndMocks.taskID,
		State: task.StateFinished,
	}).Return(task.CompletingResultCompleted, nil)

	response, err := handlerAndMocks.handler.HandleRequest(handlerAndMocks.request)
	handlerAndMocks.assertExpectations(t)
	assert.NoError(t, err)
	assert.Equal(t, internalHTTP.Response{
		StatusCode: http.StatusCreated,
		Body:       handlerAndMocks.request.Body,
		Headers:    map[string]string{api.ContentTypeHeaderName: api.ContentTypeApplicationJSON},
	}, response)
}

func TestPutTaskCompletionRequestHandler_HandleRequest_InvalidPayload(t *testing.T) {
	handler := handlers.NewPutTaskCompletionRequestHandler(new(taskCompleterMock))
	response, err := handler.HandleRequest(internalHTTP.Request{
		Body: "",
	})

	assert.NoError(t, err)
	assert.Equal(t, internalHTTP.Response{
		StatusCode: http.StatusBadRequest,
		Body:       handlers.InvalidPayloadErrorMessage,
		Headers:    map[string]string{api.ContentTypeHeaderName: api.ContentTypeTextPlain},
	}, response)
}

func TestPutTaskCompletionRequestHandler_HandleRequest_UnknownCompletionState(t *testing.T) {
	handler := handlers.NewPutTaskCompletionRequestHandler(new(taskCompleterMock))
	completion := internalHTTP.Completion{State: internalHTTP.CompletionState("invalid")}
	response, err := handler.HandleRequest(internalHTTP.Request{
		Body: completion.JSON(),
	})

	assert.NoError(t, err)
	assert.Equal(t, internalHTTP.Response{
		StatusCode: http.StatusBadRequest,
		Body:       handlers.UnknownCompletionStateMsg,
		Headers:    map[string]string{api.ContentTypeHeaderName: api.ContentTypeTextPlain},
	}, response)
}

func TestPutTaskCompletionRequestHandler_HandleRequest_CompleteWithError(t *testing.T) {
	errorMsg := "error"
	completion := internalHTTP.Completion{State: internalHTTP.CompletionStateError, ErrorMessage: &errorMsg}
	handlerAndMocks := newPutTaskCompletionReqHandlerWithMocks(completion)
	handlerAndMocks.completerMock.On("Complete", task.CompleteRequest{
		ID:      handlerAndMocks.taskID,
		State:   task.StateAborted,
		Message: &errorMsg,
	}).Return(task.CompletingResultCompleted, nil)

	response, err := handlerAndMocks.handler.HandleRequest(handlerAndMocks.request)
	handlerAndMocks.assertExpectations(t)
	assert.NoError(t, err)
	assert.Equal(t, internalHTTP.Response{
		StatusCode: http.StatusCreated,
		Body:       handlerAndMocks.request.Body,
		Headers:    map[string]string{api.ContentTypeHeaderName: api.ContentTypeApplicationJSON},
	}, response)
}

func TestPutTaskCompletionRequestHandler_HandleRequest_TaskStateConflict(t *testing.T) {
	completion := internalHTTP.Completion{State: internalHTTP.CompletionStateCompleted}
	handlerAndMocks := newPutTaskCompletionReqHandlerWithMocks(completion)
	handlerAndMocks.completerMock.On("Complete", task.CompleteRequest{
		ID:    handlerAndMocks.taskID,
		State: task.StateFinished,
	}).Return(task.CompletingResultConflict, nil)

	response, err := handlerAndMocks.handler.HandleRequest(handlerAndMocks.request)
	handlerAndMocks.assertExpectations(t)
	assert.NoError(t, err)
	assert.Equal(t, internalHTTP.Response{
		StatusCode: http.StatusConflict,
		Body:       handlers.ConflictingTaskCompletionMsg,
		Headers:    map[string]string{api.ContentTypeHeaderName: api.ContentTypeTextPlain},
	}, response)
}

func TestPutTaskCompletionRequestHandler_HandleRequest_UnknownCompletionResult(t *testing.T) {
	completion := internalHTTP.Completion{State: internalHTTP.CompletionStateCompleted}
	handlerAndMocks := newPutTaskCompletionReqHandlerWithMocks(completion)
	handlerAndMocks.completerMock.On("Complete", task.CompleteRequest{
		ID:    handlerAndMocks.taskID,
		State: task.StateFinished,
	}).Return(task.CompletingResult("unknown"), nil)

	response, err := handlerAndMocks.handler.HandleRequest(handlerAndMocks.request)
	handlerAndMocks.assertExpectations(t)
	assert.NoError(t, err)
	assert.Equal(t, internalHTTP.Response{
		StatusCode: http.StatusInternalServerError,
		Body:       handlers.UnknownErrorMsg,
		Headers:    map[string]string{api.ContentTypeHeaderName: api.ContentTypeTextPlain},
	}, response)
}

func TestPutTaskCompletionRequestHandler_HandleRequest_CompletionError(t *testing.T) {
	completion := internalHTTP.Completion{State: internalHTTP.CompletionStateCompleted}
	handlerAndMocks := newPutTaskCompletionReqHandlerWithMocks(completion)
	handlerAndMocks.completerMock.On("Complete", task.CompleteRequest{
		ID:    handlerAndMocks.taskID,
		State: task.StateFinished,
	}).Return(task.CompletingResultCompleted, errors.New("error"))

	_, err := handlerAndMocks.handler.HandleRequest(handlerAndMocks.request)
	handlerAndMocks.assertExpectations(t)
	assert.Error(t, err)
}
