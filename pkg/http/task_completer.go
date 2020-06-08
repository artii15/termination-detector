package http

import (
	"fmt"
	"net/http"

	"github.com/artii15/termination-detector/pkg/task"
)

type TaskCompleter struct {
	requestExecutor requestExecutor
}

func NewTaskCompleter(requestExecutor requestExecutor) *TaskCompleter {
	return &TaskCompleter{
		requestExecutor: requestExecutor,
	}
}

func (completer *TaskCompleter) Complete(request task.CompleteRequest) (task.CompletingResult, error) {
	completionState, isCompletionStateDefined := taskStateToCompletionStateMapping[request.State]
	if !isCompletionStateDefined {
		return "", fmt.Errorf("not allowed terminal task state requested: %s", request.State)
	}
	taskCompletion := Completion{
		State:        completionState,
		ErrorMessage: request.Message,
	}
	response, err := completer.requestExecutor.ExecuteRequest(Request{
		Method:       MethodPut,
		ResourcePath: ResourcePathTaskCompletion,
		Body:         taskCompletion.JSON(),
		PathParameters: map[PathParameter]string{
			PathParameterProcessID: request.ProcessID,
			PathParameterTaskID:    request.TaskID,
		},
	})
	if err != nil {
		return "", err
	}
	switch response.StatusCode {
	case http.StatusCreated:
		return task.CompletingResultCompleted, nil
	case http.StatusConflict:
		return task.CompletingResultConflict, nil
	default:
		return "", fmt.Errorf("unexpected completion result: %d %s", response.StatusCode, response.Body)
	}
}
