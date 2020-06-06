package dynamo

import (
	"fmt"
	"time"

	task2 "github.com/nordcloud/termination-detector/pkg/task"

	"github.com/aws/aws-sdk-go/service/dynamodb"
)

const (
	ProcessIDAttrName             = "process_id"
	TaskIDAttrName                = "task_id"
	TaskBadStateEnterTimeAttrName = "bad_state_enter_time"
	TaskStateAttrName             = "state"
	TaskStateMessageAttrName      = "state_message"
	taskExpirationTimeAttrName    = "expiration_time"
	taskTTLAttributeName          = "ttl"

	ProcessIDAttrAlias             = "#processID"
	taskIDAttrAlias                = "#taskID"
	taskBadStateEnterTimeAttrAlias = "#badStateEnterTime"
	taskStateAttrAlias             = "#state"
	taskExpirationTimeAttrAlias    = "#expirationTime"
	taskTTLAttrAlias               = "#ttl"
	taskStateMessageAttrAlias      = "#stateMessage"

	ProcessIDValuePlaceholder             = ":processID"
	taskStateCreatedValuePlaceholder      = ":stateCreated"
	taskBadStateEnterTimeValuePlaceholder = ":badStateEnterTime"

	taskBadStateEnterTimeZeroValue = "0"
)

func readTaskBadStateEnterTime(dynamoTask map[string]*dynamodb.AttributeValue) (time.Time, error) {
	badStateEnterTimeAttr, isBadStateEnterTimeDefined := dynamoTask[TaskBadStateEnterTimeAttrName]
	if !isBadStateEnterTimeDefined || badStateEnterTimeAttr.S == nil {
		return time.Time{}, fmt.Errorf("item does not contain bad state enter time attribute: %+v", dynamoTask)
	}
	badStateEnterTime, err := time.Parse(time.RFC3339, *badStateEnterTimeAttr.S)
	if err != nil {
		return time.Time{}, err
	}
	return badStateEnterTime, nil
}

func readTaskState(dynamoTask map[string]*dynamodb.AttributeValue) (task2.State, error) {
	taskStateAttr, isTaskStateDefined := dynamoTask[TaskStateAttrName]
	if !isTaskStateDefined || taskStateAttr.S == nil {
		return "", fmt.Errorf("item does not contain task state attribute: %+v", dynamoTask)
	}
	return task2.State(*taskStateAttr.S), nil
}

func readTaskStateMessage(dynamoTask map[string]*dynamodb.AttributeValue) *string {
	stateMsgAttr, isStateMsgDefined := dynamoTask[TaskStateMessageAttrName]
	if !isStateMsgDefined || stateMsgAttr.S == nil {
		return nil
	}
	return stateMsgAttr.S
}
