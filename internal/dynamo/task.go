package dynamo

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	internalTask "github.com/nordcloud/termination-detector/internal/task"
)

type task struct {
	TaskID          string             `json:"task_id"`
	ProcessID       string             `json:"process_id"`
	ExpirationTime  time.Time          `json:"expiration_time"`
	State           internalTask.State `json:"state"`
	TTL             int64              `json:"ttl"`
	ProcessingState string             `json:"processing_state"`
	StateMessage    *string            `json:"state_message"`
}

func (task task) dynamoItem() map[string]*dynamodb.AttributeValue {
	marshalled, err := dynamodbattribute.MarshalMap(task)
	if err != nil {
		panic(fmt.Sprintf("failed to marshal task: %s\n%+v", err.Error(), task))
	}
	return marshalled
}

func (task task) ID() internalTask.ID {
	return internalTask.ID{
		TaskID:    task.TaskID,
		ProcessID: task.ProcessID,
	}
}

func newTask(toConvert internalTask.Task, ttl int64) task {
	return task{
		TaskID:          toConvert.TaskID,
		ProcessID:       toConvert.ProcessID,
		ExpirationTime:  toConvert.ExpirationTime,
		State:           toConvert.State,
		StateMessage:    toConvert.StateMessage,
		TTL:             ttl,
		ProcessingState: makeProcessingState(toConvert.State, toConvert.ExpirationTime),
	}
}

func makeProcessingState(state internalTask.State, expirationTime time.Time) string {
	return fmt.Sprintf("%s__%s", state, expirationTime)
}

func readDynamoTask(dynamoTask map[string]*dynamodb.AttributeValue) (unmarshalled task) {
	if err := dynamodbattribute.UnmarshalMap(dynamoTask, &unmarshalled); err != nil {
		panic(fmt.Sprintf("failed to unmarshal task: %s\n%+v", err.Error(), dynamoTask))
	}
	return unmarshalled
}
