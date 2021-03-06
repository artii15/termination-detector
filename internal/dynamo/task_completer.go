package dynamo

import (
	"fmt"
	"time"

	"github.com/artii15/termination-detector/pkg/task"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
)

const (
	currentTimeValuePlaceholder         = ":currentTime"
	newTaskStateValuePlaceholder        = ":newState"
	newTaskStateMessageValuePlaceholder = ":newStateMessage"
)

var (
	completeTaskUpdateExpr = fmt.Sprintf("SET %s = %s, %s = %s, %s = %s",
		taskStateAttrAlias, newTaskStateValuePlaceholder,
		taskStateMessageAttrAlias, newTaskStateMessageValuePlaceholder,
		taskBadStateEnterTimeAttrAlias, taskBadStateEnterTimeValuePlaceholder)
	completeTaskConditionExpr = fmt.Sprintf("attribute_exists(%s) and attribute_exists(%s) and %s > %s and %s = %s",
		ProcessIDAttrAlias, taskIDAttrAlias, taskExpirationTimeAttrAlias, currentTimeValuePlaceholder,
		taskStateAttrAlias, taskStateCreatedValuePlaceholder)
)

type TaskCompleter struct {
	dynamoAPI         dynamodbiface.DynamoDBAPI
	tasksTableName    string
	currentDateGetter currentDateGetter
}

func NewTaskCompleter(dynamoAPI dynamodbiface.DynamoDBAPI, tasksTableName string,
	currentDateGetter currentDateGetter) *TaskCompleter {
	return &TaskCompleter{
		dynamoAPI:         dynamoAPI,
		tasksTableName:    tasksTableName,
		currentDateGetter: currentDateGetter,
	}
}

func (completer *TaskCompleter) Complete(request task.CompleteRequest) (task.CompletingResult, error) {
	updateItemInput := BuildCompleteTaskUpdateItemInput(completer.tasksTableName, CompleteTaskRequest{
		CompletionTime: completer.currentDateGetter.GetCurrentDate(),
		TerminalState:  request.State,
		Message:        request.Message,
		ProcessID:      request.ProcessID,
		TaskID:         request.TaskID,
	})
	_, err := completer.dynamoAPI.UpdateItem(updateItemInput)
	if err != nil {
		if awsErr, isAWSErr := err.(awserr.Error); isAWSErr && awsErr.Code() == dynamodb.ErrCodeConditionalCheckFailedException {
			return task.CompletingResultConflict, nil
		}
		return "", err
	}
	return task.CompletingResultCompleted, nil
}

type CompleteTaskRequest struct {
	CompletionTime time.Time
	TerminalState  task.State
	Message        *string
	ProcessID      string
	TaskID         string
}

func BuildCompleteTaskUpdateItemInput(tableName string, completeTaskRequest CompleteTaskRequest) *dynamodb.UpdateItemInput {
	expressionAttributeValues := map[string]*dynamodb.AttributeValue{
		currentTimeValuePlaceholder:           {S: aws.String(completeTaskRequest.CompletionTime.Format(time.RFC3339))},
		taskStateCreatedValuePlaceholder:      {S: aws.String(string(task.StateCreated))},
		newTaskStateValuePlaceholder:          {S: aws.String(string(completeTaskRequest.TerminalState))},
		newTaskStateMessageValuePlaceholder:   {NULL: aws.Bool(true)},
		taskBadStateEnterTimeValuePlaceholder: {S: aws.String(taskBadStateEnterTimeZeroValue)},
	}
	if completeTaskRequest.Message != nil {
		expressionAttributeValues[newTaskStateMessageValuePlaceholder] = &dynamodb.AttributeValue{S: completeTaskRequest.Message}
	}
	if completeTaskRequest.TerminalState == task.StateAborted {
		expressionAttributeValues[taskBadStateEnterTimeValuePlaceholder] = &dynamodb.AttributeValue{S: aws.String(completeTaskRequest.CompletionTime.Format(time.RFC3339))}
	}

	return &dynamodb.UpdateItemInput{
		ConditionExpression: &completeTaskConditionExpr,
		ExpressionAttributeNames: map[string]*string{
			ProcessIDAttrAlias:             aws.String(ProcessIDAttrName),
			taskIDAttrAlias:                aws.String(TaskIDAttrName),
			taskExpirationTimeAttrAlias:    aws.String(taskExpirationTimeAttrName),
			taskStateAttrAlias:             aws.String(TaskStateAttrName),
			taskStateMessageAttrAlias:      aws.String(TaskStateMessageAttrName),
			taskBadStateEnterTimeAttrAlias: aws.String(TaskBadStateEnterTimeAttrName),
		},
		ExpressionAttributeValues: expressionAttributeValues,
		Key: map[string]*dynamodb.AttributeValue{
			ProcessIDAttrName: {S: &completeTaskRequest.ProcessID},
			TaskIDAttrName:    {S: &completeTaskRequest.TaskID},
		},
		TableName:        &tableName,
		UpdateExpression: &completeTaskUpdateExpr,
	}
}
