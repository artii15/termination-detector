package dynamo

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
	"github.com/nordcloud/termination-detector/internal/process"
	"github.com/nordcloud/termination-detector/internal/task"
)

const (
	taskBadStateEnterTimeIndex = "badStateEnterTimeIndex"
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
	out, err := getter.dynamoAPI.Query(&dynamodb.QueryInput{
		ConsistentRead: aws.Bool(true),
		ExpressionAttributeNames: map[string]*string{
			"#processID": aws.String("process_id"),
		},
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":processID": {S: &processID},
		},
		KeyConditionExpression: aws.String("#processID = :processID"),
		Limit:                  aws.Int64(1),
		TableName:              &getter.tasksTableName,
	})
	if err != nil {
		return false, err
	}
	return out != nil && len(out.Items) > 0, nil
}

func (getter *ProcessGetter) getProcess(processID string) (process.Process, error) {
	out, err := getter.dynamoAPI.Query(&dynamodb.QueryInput{
		ConsistentRead: aws.Bool(true),
		ExpressionAttributeNames: map[string]*string{
			"#processID":         aws.String("process_id"),
			"#badStateEnterTime": aws.String("bad_state_enter_time"),
		},
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":processID":                  {S: &processID},
			":badStateEnterTimeZeroValue": {S: aws.String(taskBadStateEnterTimeZeroValue)},
		},
		IndexName:              aws.String(taskBadStateEnterTimeIndex),
		KeyConditionExpression: aws.String("#processID = :processID and #badStateEnterTime > :badStateEnterTimeZeroValue"),
		Limit:                  aws.Int64(1),
		TableName:              &getter.tasksTableName,
	})
	if err != nil {
		return process.Process{}, err
	}
	if out == nil || len(out.Items) == 0 {
		return process.Process{ID: processID, State: process.StateCompleted}, nil
	}

	dynamoTask := out.Items[0]
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
	badStateEnterTime, err := readTaskBadStateEnterTime(dynamoTask)
	if err != nil {
		return process.Process{}, err
	}

	currentDate := getter.currentDateGetter.GetCurrentDate()
	if badStateEnterTime.After(currentDate) {
		return process.Process{
			ID:    processID,
			State: process.StateCreated,
		}, nil
	}

	return process.Process{
		ID:           processID,
		State:        process.StateError,
		StateMessage: aws.String("process timed out"),
	}, nil
}
