package handlers

import (
	"net/http"

	internalHTTP "github.com/nordcloud/termination-detector/pkg/http"
	"github.com/nordcloud/termination-detector/pkg/task"
	"github.com/sirupsen/logrus"
)

const (
	UnknownCompletionStateMsg    = "unknown completion state"
	ConflictingTaskCompletionMsg = "task not created or already completed"
	UnknownErrorMsg              = "unknown error"
)

var completionStateToTaskStateMapping = map[internalHTTP.CompletionState]task.State{
	internalHTTP.CompletionStateError:     task.StateAborted,
	internalHTTP.CompletionStateCompleted: task.StateFinished,
}

type PutTaskCompletionRequestHandler struct {
	completer task.Completer
}

func NewPutTaskCompletionRequestHandler(completer task.Completer) *PutTaskCompletionRequestHandler {
	return &PutTaskCompletionRequestHandler{
		completer: completer,
	}
}

func (handler *PutTaskCompletionRequestHandler) HandleRequest(request internalHTTP.Request) (internalHTTP.Response, error) {
	completion, err := internalHTTP.UnmarshalCompletion(request.Body)
	if err != nil {
		return internalHTTP.Response{
			StatusCode: http.StatusBadRequest,
			Body:       InvalidPayloadErrorMessage,
			Headers:    map[string]string{internalHTTP.ContentTypeHeaderName: internalHTTP.ContentTypeTextPlain},
		}, nil
	}
	taskCompletionState, isTaskCompletionStateFound := completionStateToTaskStateMapping[completion.State]
	if !isTaskCompletionStateFound {
		return internalHTTP.Response{
			StatusCode: http.StatusBadRequest,
			Body:       UnknownCompletionStateMsg,
			Headers: map[string]string{
				internalHTTP.ContentTypeHeaderName: internalHTTP.ContentTypeTextPlain,
			},
		}, nil
	}

	completingResult, err := handler.completer.Complete(task.CompleteRequest{
		ID: task.ID{
			ProcessID: request.PathParameters[internalHTTP.PathParameterProcessID],
			TaskID:    request.PathParameters[internalHTTP.PathParameterTaskID],
		},
		State:   taskCompletionState,
		Message: completion.ErrorMessage,
	})
	if err != nil {
		return internalHTTP.Response{}, err
	}

	return mapCompletingResultToResponse(request, completingResult), nil
}

func mapCompletingResultToResponse(request internalHTTP.Request, result task.CompletingResult) internalHTTP.Response {
	switch result {
	case task.CompletingResultConflict:
		return internalHTTP.Response{
			StatusCode: http.StatusConflict,
			Body:       ConflictingTaskCompletionMsg,
			Headers: map[string]string{
				internalHTTP.ContentTypeHeaderName: internalHTTP.ContentTypeTextPlain,
			},
		}
	case task.CompletingResultCompleted:
		return internalHTTP.Response{
			StatusCode: http.StatusCreated,
			Body:       request.Body,
			Headers: map[string]string{
				internalHTTP.ContentTypeHeaderName: internalHTTP.ContentTypeApplicationJSON,
			},
		}
	default:
		logrus.WithField("unknown_completion_result", result).Error("unknown task completion result")
		return internalHTTP.Response{
			StatusCode: http.StatusInternalServerError,
			Body:       UnknownErrorMsg,
			Headers: map[string]string{
				internalHTTP.ContentTypeHeaderName: internalHTTP.ContentTypeTextPlain,
			},
		}
	}
}
