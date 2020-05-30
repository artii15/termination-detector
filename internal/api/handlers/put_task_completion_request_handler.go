package handlers

import (
	"net/http"

	task2 "github.com/nordcloud/termination-detector/pkg/task"

	"github.com/nordcloud/termination-detector/internal/api"
	"github.com/sirupsen/logrus"
)

const (
	UnknownCompletionStateMsg    = "unknown completion state"
	ConflictingTaskCompletionMsg = "task not created or already completed"
	UnknownErrorMsg              = "unknown error"
)

var completionStateToTaskStateMapping = map[api.CompletionState]task2.State{
	api.CompletionStateError:     task2.StateAborted,
	api.CompletionStateCompleted: task2.StateFinished,
}

type PutTaskCompletionRequestHandler struct {
	completer task2.Completer
}

func NewPutTaskCompletionRequestHandler(completer task2.Completer) *PutTaskCompletionRequestHandler {
	return &PutTaskCompletionRequestHandler{
		completer: completer,
	}
}

func (handler *PutTaskCompletionRequestHandler) HandleRequest(request api.Request) (api.Response, error) {
	completion, err := api.UnmarshalCompletion(request.Body)
	if err != nil {
		return api.Response{
			StatusCode: http.StatusBadRequest,
			Body:       InvalidPayloadErrorMessage,
			Headers:    map[string]string{api.ContentTypeHeaderName: api.ContentTypeTextPlain},
		}, nil
	}
	taskCompletionState, isTaskCompletionStateFound := completionStateToTaskStateMapping[completion.State]
	if !isTaskCompletionStateFound {
		return api.Response{
			StatusCode: http.StatusBadRequest,
			Body:       UnknownCompletionStateMsg,
			Headers: map[string]string{
				api.ContentTypeHeaderName: api.ContentTypeTextPlain,
			},
		}, nil
	}

	completingResult, err := handler.completer.Complete(task2.CompleteRequest{
		ID: task2.ID{
			ProcessID: request.PathParameters[api.ProcessIDPathParameter],
			TaskID:    request.PathParameters[api.TaskIDPathParameter],
		},
		State:   taskCompletionState,
		Message: completion.ErrorMessage,
	})
	if err != nil {
		return api.Response{}, err
	}

	return mapCompletingResultToResponse(request, completingResult), nil
}

func mapCompletingResultToResponse(request api.Request, result task2.CompletingResult) api.Response {
	switch result {
	case task2.CompletingResultConflict:
		return api.Response{
			StatusCode: http.StatusConflict,
			Body:       ConflictingTaskCompletionMsg,
			Headers: map[string]string{
				api.ContentTypeHeaderName: api.ContentTypeTextPlain,
			},
		}
	case task2.CompletingResultCompleted:
		return api.Response{
			StatusCode: http.StatusCreated,
			Body:       request.Body,
			Headers: map[string]string{
				api.ContentTypeHeaderName: api.ContentTypeApplicationJSON,
			},
		}
	default:
		logrus.WithField("unknown_completion_result", result).Error("unknown task completion result")
		return api.Response{
			StatusCode: http.StatusInternalServerError,
			Body:       UnknownErrorMsg,
			Headers: map[string]string{
				api.ContentTypeHeaderName: api.ContentTypeTextPlain,
			},
		}
	}
}
