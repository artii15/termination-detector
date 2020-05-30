package dynamo

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
	"github.com/nordcloud/termination-detector/internal/process"
	"github.com/nordcloud/termination-detector/internal/task"
)

const (
	taskBadStateEnterTimeIndex                = "badStateEnterTimeIndex"
	taskBadStateEnterTimeZeroValuePlaceholder = ":badStateEnterTimeZeroValue"
)

var (
	queryProcExistsKeyCondExpression = fmt.Sprintf("%s = %s", processIDAttrAlias, processIDValuePlaceholder)
	queryGetProcessKeyCondExpression = fmt.Sprintf("%s = %s and %s > %s", processIDAttrAlias,
		processIDValuePlaceholder, taskBadStateEnterTimeAttrAlias, taskBadStateEnterTimeZeroValuePlaceholder)
)

type ProcessGetter struct {
	dynamoAPI         dynamodbiface.DynamoDBAPI
	tasksTableName    string
	currentDateGetter currentDateGetter
}

func NewProcessGetter(dynamoAPI dynamodbiface.DynamoDBAPI,
	tasksTableName string, currentDateGetter currentDateGetter) *ProcessGetter {
	return &ProcessGetter{
		dynamoAPI:         dynamoAPI,
		tasksTableName:    tasksTableName,
		currentDateGetter: currentDateGetter,
	}
}

func (getter *ProcessGetter) Get(processID string) (*process.Process, error) {
	if processExists, err := getter.exists(processID); err != nil || !processExists {
		return nil, err
	}

	foundProcess, err := getter.getProcess(processID)
	return &foundProcess, err
}

func (getter *ProcessGetter) exists(processID string) (bool, error) {
	out, err := getter.dynamoAPI.Query(BuildCheckIfProcessExistsQueryInput(getter.tasksTableName, processID))
	if err != nil {
		return false, err
	}
	return out != nil && len(out.Items) > 0, nil
}

func BuildCheckIfProcessExistsQueryInput(tableName, processID string) *dynamodb.QueryInput {
	return &dynamodb.QueryInput{
		ConsistentRead: aws.Bool(true),
		ExpressionAttributeNames: map[string]*string{
			processIDAttrAlias: aws.String(ProcessIDAttrName),
		},
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			processIDValuePlaceholder: {S: &processID},
		},
		KeyConditionExpression: &queryProcExistsKeyCondExpression,
		Limit:                  aws.Int64(1),
		TableName:              &tableName,
	}
}

func (getter *ProcessGetter) getProcess(processID string) (process.Process, error) {
	queryResult, err := getter.dynamoAPI.Query(BuildGetProcessQueryInput(getter.tasksTableName, processID))
	if err != nil {
		return process.Process{}, err
	}
	if queryResult == nil || len(queryResult.Items) == 0 {
		return process.Process{ID: processID, State: process.StateCompleted}, nil
	}
	return getter.readNotCompletedProcess(processID, queryResult.Items[0])
}

func (getter *ProcessGetter) readNotCompletedProcess(processID string, dynamoTask map[string]*dynamodb.AttributeValue) (
	process.Process, error) {
	taskState, err := readTaskState(dynamoTask)
	if err != nil {
		return process.Process{}, err
	}
	if taskState == task.StateAborted {
		return process.Process{
			ID:           processID,
			State:        process.StateError,
			StateMessage: readTaskStateMessage(dynamoTask),
		}, nil
	}
	if taskState != task.StateCreated {
		return process.Process{}, fmt.Errorf("unexpected task state: %+v", dynamoTask)
	}

	return getter.reportFailureIfTaskTimedOut(processID, dynamoTask)
}

func (getter *ProcessGetter) reportFailureIfTaskTimedOut(processID string, dynamoTask map[string]*dynamodb.AttributeValue) (process.Process, error) {
	badStateEnterTime, err := readTaskBadStateEnterTime(dynamoTask)
	if err != nil {
		return process.Process{}, err
	}
	if getter.currentDateGetter.GetCurrentDate().Before(badStateEnterTime) {
		return process.Process{
			ID:    processID,
			State: process.StateCreated,
		}, nil
	}
	return process.Process{
		ID:           processID,
		State:        process.StateError,
		StateMessage: aws.String(process.TimedOutErrorMessage),
	}, nil
}

func BuildGetProcessQueryInput(tableName, processID string) *dynamodb.QueryInput {
	return &dynamodb.QueryInput{
		ConsistentRead: aws.Bool(true),
		ExpressionAttributeNames: map[string]*string{
			processIDAttrAlias:             aws.String(processIDValuePlaceholder),
			taskBadStateEnterTimeAttrAlias: aws.String(TaskBadStateEnterTimeAttrName),
		},
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			processIDValuePlaceholder:                 {S: &processID},
			taskBadStateEnterTimeZeroValuePlaceholder: {S: aws.String(taskBadStateEnterTimeZeroValue)},
		},
		IndexName:              aws.String(taskBadStateEnterTimeIndex),
		KeyConditionExpression: &queryGetProcessKeyCondExpression,
		Limit:                  aws.Int64(1),
		TableName:              &tableName,
	}
}
