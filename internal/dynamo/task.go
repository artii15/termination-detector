package dynamo

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"

	"github.com/aws/aws-sdk-go/service/dynamodb"

	internalTask "github.com/nordcloud/termination-detector/internal/task"
)

type task struct {
	ID              string             `json:"id"`
	ProcessID       string             `json:"process_id"`
	ExpirationTime  time.Time          `json:"expiration_time"`
	State           internalTask.State `json:"state"`
	TTL             int64              `json:"ttl"`
	ProcessingState string             `json:"processing_state"`
}

func (task task) dynamoItem() map[string]*dynamodb.AttributeValue {
	marshalled, err := dynamodbattribute.MarshalMap(task)
	if err != nil {
		panic(fmt.Sprintf("failed to marshal task: %s\n%+v", err.Error(), task))
	}
	return marshalled
}

func (task task) registrationData() internalTask.RegistrationData {
	return internalTask.RegistrationData{
		ID:             task.ID,
		ProcessID:      task.ProcessID,
		ExpirationTime: task.ExpirationTime,
	}
}

func newTask(toConvert internalTask.Task, ttl int64) task {
	return task{
		ID:              toConvert.ID,
		ProcessID:       toConvert.ProcessID,
		ExpirationTime:  toConvert.ExpirationTime,
		State:           toConvert.State,
		TTL:             ttl,
		ProcessingState: fmt.Sprintf("%s__%s__%s", toConvert.State, toConvert.ExpirationTime, toConvert.ID),
	}
}

func readDynamoTask(dynamoTask map[string]*dynamodb.AttributeValue) (unmarshalled task) {
	if err := dynamodbattribute.UnmarshalMap(dynamoTask, &unmarshalled); err != nil {
		panic(fmt.Sprintf("failed to unmarshal task: %s\n%+v", err.Error(), dynamoTask))
	}
	return unmarshalled
}
