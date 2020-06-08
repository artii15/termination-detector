package http_test

import (
	"errors"
	"net/http"
	"testing"

	internalHTTP "github.com/artii15/termination-detector/pkg/http"
	"github.com/artii15/termination-detector/pkg/task"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/stretchr/testify/assert"
)

type taskCompleterWithMocks struct {
	requestExecutor *requestExecutorMock
	taskCompleter   *internalHTTP.TaskCompleter
}

func newTaskCompleterWithMocks() *taskCompleterWithMocks {
	requestExecutor := new(requestExecutorMock)
	return &taskCompleterWithMocks{
		requestExecutor: requestExecutor,
		taskCompleter:   internalHTTP.NewTaskCompleter(requestExecutor),
	}
}

func TestTaskCompleter_Complete(t *testing.T) {
	completerAndMocks := newTaskCompleterWithMocks()
	taskCompletion := internalHTTP.Completion{
		State:        internalHTTP.CompletionStateError,
		ErrorMessage: aws.String("error"),
	}
	procID := "1"
	taskID := "2"
	completerAndMocks.requestExecutor.On("ExecuteRequest", internalHTTP.Request{
		Method:       internalHTTP.MethodPut,
		ResourcePath: internalHTTP.ResourcePathTaskCompletion,
		Body:         taskCompletion.JSON(),
		PathParameters: map[internalHTTP.PathParameter]string{
			internalHTTP.PathParameterProcessID: procID,
			internalHTTP.PathParameterTaskID:    taskID,
		},
	}).Return(internalHTTP.Response{
		StatusCode: http.StatusCreated,
		Headers: map[string]string{
			internalHTTP.ContentTypeHeaderName: internalHTTP.ContentTypeApplicationJSON,
		},
	}, nil)

	completion, err := completerAndMocks.taskCompleter.Complete(task.CompleteRequest{
		ID: task.ID{
			ProcessID: procID,
			TaskID:    taskID,
		},
		State:   task.StateAborted,
		Message: taskCompletion.ErrorMessage,
	})
	assert.NoError(t, err)
	assert.Equal(t, task.CompletingResultCompleted, completion)
}

func TestTaskCompleter_Complete_ConflictingState(t *testing.T) {
	completerAndMocks := newTaskCompleterWithMocks()
	taskCompletion := internalHTTP.Completion{
		State:        internalHTTP.CompletionStateError,
		ErrorMessage: aws.String("error"),
	}
	procID := "1"
	taskID := "2"
	completerAndMocks.requestExecutor.On("ExecuteRequest", internalHTTP.Request{
		Method:       internalHTTP.MethodPut,
		ResourcePath: internalHTTP.ResourcePathTaskCompletion,
		Body:         taskCompletion.JSON(),
		PathParameters: map[internalHTTP.PathParameter]string{
			internalHTTP.PathParameterProcessID: procID,
			internalHTTP.PathParameterTaskID:    taskID,
		},
	}).Return(internalHTTP.Response{
		StatusCode: http.StatusConflict,
		Headers: map[string]string{
			internalHTTP.ContentTypeHeaderName: internalHTTP.ContentTypeTextPlain,
		},
	}, nil)

	completion, err := completerAndMocks.taskCompleter.Complete(task.CompleteRequest{
		ID: task.ID{
			ProcessID: procID,
			TaskID:    taskID,
		},
		State:   task.StateAborted,
		Message: taskCompletion.ErrorMessage,
	})
	assert.NoError(t, err)
	assert.Equal(t, task.CompletingResultConflict, completion)
}

func TestTaskCompleter_Complete_UnexpectedTaskState(t *testing.T) {
	completerAndMocks := newTaskCompleterWithMocks()
	taskCompletion := internalHTTP.Completion{
		State:        internalHTTP.CompletionStateError,
		ErrorMessage: aws.String("error"),
	}
	procID := "1"
	taskID := "2"
	completerAndMocks.requestExecutor.On("ExecuteRequest", internalHTTP.Request{
		Method:       internalHTTP.MethodPut,
		ResourcePath: internalHTTP.ResourcePathTaskCompletion,
		Body:         taskCompletion.JSON(),
		PathParameters: map[internalHTTP.PathParameter]string{
			internalHTTP.PathParameterProcessID: procID,
			internalHTTP.PathParameterTaskID:    taskID,
		},
	}).Return(internalHTTP.Response{
		StatusCode: http.StatusBadRequest,
		Headers: map[string]string{
			internalHTTP.ContentTypeHeaderName: internalHTTP.ContentTypeTextPlain,
		},
	}, nil)

	_, err := completerAndMocks.taskCompleter.Complete(task.CompleteRequest{
		ID: task.ID{
			ProcessID: procID,
			TaskID:    taskID,
		},
		State:   task.StateAborted,
		Message: taskCompletion.ErrorMessage,
	})
	assert.Error(t, err)
}

func TestTaskCompleter_Complete_ExecutorError(t *testing.T) {
	completerAndMocks := newTaskCompleterWithMocks()
	taskCompletion := internalHTTP.Completion{
		State:        internalHTTP.CompletionStateError,
		ErrorMessage: aws.String("error"),
	}
	procID := "1"
	taskID := "2"
	completerAndMocks.requestExecutor.On("ExecuteRequest", internalHTTP.Request{
		Method:       internalHTTP.MethodPut,
		ResourcePath: internalHTTP.ResourcePathTaskCompletion,
		Body:         taskCompletion.JSON(),
		PathParameters: map[internalHTTP.PathParameter]string{
			internalHTTP.PathParameterProcessID: procID,
			internalHTTP.PathParameterTaskID:    taskID,
		},
	}).Return(internalHTTP.Response{
		StatusCode: http.StatusBadRequest,
		Headers: map[string]string{
			internalHTTP.ContentTypeHeaderName: internalHTTP.ContentTypeTextPlain,
		},
	}, errors.New("error"))

	_, err := completerAndMocks.taskCompleter.Complete(task.CompleteRequest{
		ID: task.ID{
			ProcessID: procID,
			TaskID:    taskID,
		},
		State:   task.StateAborted,
		Message: taskCompletion.ErrorMessage,
	})
	assert.Error(t, err)
}

func TestTaskCompleter_Complete_InvalidCompletionStateRequested(t *testing.T) {
	completerAndMocks := newTaskCompleterWithMocks()
	procID := "1"
	taskID := "2"

	_, err := completerAndMocks.taskCompleter.Complete(task.CompleteRequest{
		ID: task.ID{
			ProcessID: procID,
			TaskID:    taskID,
		},
		State:   "unknown",
		Message: aws.String("failure"),
	})
	assert.Error(t, err)
}
