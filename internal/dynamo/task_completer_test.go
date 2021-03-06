package dynamo_test

import (
	"errors"
	"testing"
	"time"

	"github.com/artii15/termination-detector/internal/dynamo"
	"github.com/artii15/termination-detector/pkg/task"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/stretchr/testify/assert"
)

type taskCompleterWithMocks struct {
	completer         *dynamo.TaskCompleter
	dynamoAPI         *dynamoAPIMock
	currentDateGetter *currentDateGetterMock
}

func (completerAndMocks *taskCompleterWithMocks) assertExpectations(t *testing.T) {
	completerAndMocks.dynamoAPI.AssertExpectations(t)
	completerAndMocks.currentDateGetter.AssertExpectations(t)
}

func newTaskCompleterWithMocks() *taskCompleterWithMocks {
	dynamoAPI := new(dynamoAPIMock)
	currentDateGetter := new(currentDateGetterMock)
	return &taskCompleterWithMocks{
		completer:         dynamo.NewTaskCompleter(dynamoAPI, tasksTableName, currentDateGetter),
		dynamoAPI:         dynamoAPI,
		currentDateGetter: currentDateGetter,
	}
}

func TestTaskCompleter_Complete(t *testing.T) {
	completerAndMocks := newTaskCompleterWithMocks()
	completeTaskRequest := task.CompleteRequest{
		ID: task.ID{
			ProcessID: "2",
			TaskID:    "1",
		},
		State:   task.StateAborted,
		Message: aws.String("failed to execute task"),
	}
	completionTime := time.Now().UTC()
	completerAndMocks.currentDateGetter.On("GetCurrentDate").Return(completionTime)
	updateItemInput := dynamo.BuildCompleteTaskUpdateItemInput(tasksTableName, dynamo.CompleteTaskRequest{
		CompletionTime: completionTime,
		TerminalState:  completeTaskRequest.State,
		Message:        completeTaskRequest.Message,
		ProcessID:      completeTaskRequest.ProcessID,
		TaskID:         completeTaskRequest.TaskID,
	})
	completerAndMocks.dynamoAPI.On("UpdateItem", updateItemInput).Return(&dynamodb.UpdateItemOutput{}, nil)

	taskCompletionResult, err := completerAndMocks.completer.Complete(completeTaskRequest)
	assert.NoError(t, err)
	completerAndMocks.assertExpectations(t)
	assert.Equal(t, task.CompletingResultCompleted, taskCompletionResult)
}

func TestTaskCompleter_Complete_TaskAlreadyCompleted(t *testing.T) {
	completerAndMocks := newTaskCompleterWithMocks()
	completeTaskRequest := task.CompleteRequest{
		ID: task.ID{
			ProcessID: "2",
			TaskID:    "1",
		},
		State:   task.StateAborted,
		Message: aws.String("failed to execute task"),
	}
	completionTime := time.Now().UTC()
	completerAndMocks.currentDateGetter.On("GetCurrentDate").Return(completionTime)
	updateItemInput := dynamo.BuildCompleteTaskUpdateItemInput(tasksTableName, dynamo.CompleteTaskRequest{
		CompletionTime: completionTime,
		TerminalState:  completeTaskRequest.State,
		Message:        completeTaskRequest.Message,
		ProcessID:      completeTaskRequest.ProcessID,
		TaskID:         completeTaskRequest.TaskID,
	})
	updateErr := awserr.New(dynamodb.ErrCodeConditionalCheckFailedException, "", nil)
	completerAndMocks.dynamoAPI.On("UpdateItem", updateItemInput).Return(&dynamodb.UpdateItemOutput{}, updateErr)

	taskCompletionResult, err := completerAndMocks.completer.Complete(completeTaskRequest)
	assert.NoError(t, err)
	completerAndMocks.assertExpectations(t)
	assert.Equal(t, task.CompletingResultConflict, taskCompletionResult)
}

func TestTaskCompleter_Complete_UnexpectedError(t *testing.T) {
	completerAndMocks := newTaskCompleterWithMocks()
	completeTaskRequest := task.CompleteRequest{
		ID: task.ID{
			ProcessID: "2",
			TaskID:    "1",
		},
		State:   task.StateAborted,
		Message: aws.String("failed to execute task"),
	}
	completionTime := time.Now().UTC()
	completerAndMocks.currentDateGetter.On("GetCurrentDate").Return(completionTime)
	updateItemInput := dynamo.BuildCompleteTaskUpdateItemInput(tasksTableName, dynamo.CompleteTaskRequest{
		CompletionTime: completionTime,
		TerminalState:  completeTaskRequest.State,
		Message:        completeTaskRequest.Message,
		ProcessID:      completeTaskRequest.ProcessID,
		TaskID:         completeTaskRequest.TaskID,
	})
	completerAndMocks.dynamoAPI.On("UpdateItem", updateItemInput).Return(&dynamodb.UpdateItemOutput{}, errors.New("error"))

	_, err := completerAndMocks.completer.Complete(completeTaskRequest)
	assert.Error(t, err)
	completerAndMocks.assertExpectations(t)
}
