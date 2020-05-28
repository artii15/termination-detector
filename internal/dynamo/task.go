package dynamo

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/nordcloud/termination-detector/internal/task"
)

func readTaskBadStateEnterTime(dynamoTask map[string]*dynamodb.AttributeValue) (time.Time, error) {
	badStateEnterTimeAttr, isBadStateEnterTimeDefined := dynamoTask["bad_state_enter_time"]
	if !isBadStateEnterTimeDefined || badStateEnterTimeAttr.S == nil {
		return time.Time{}, fmt.Errorf("item does not contain bad state enter time attribute: %+v", dynamoTask)
	}
	badStateEnterTime, err := time.Parse(time.RFC3339, *badStateEnterTimeAttr.S)
	if err != nil {
		return time.Time{}, err
	}
	return badStateEnterTime, nil
}

func readTaskState(dynamoTask map[string]*dynamodb.AttributeValue) (task.State, error) {
	taskStateAttr, isTaskStateDefined := dynamoTask["state"]
	if !isTaskStateDefined || taskStateAttr.S == nil {
		return "", fmt.Errorf("item does not contain task state attribute: %+v", dynamoTask)
	}
	return task.State(*taskStateAttr.S), nil
}

func readTaskStateMessage(dynamoTask map[string]*dynamodb.AttributeValue) *string {
	stateMsgAttr, isStateMsgDefined := dynamoTask["state_message"]
	if !isStateMsgDefined || stateMsgAttr.S == nil {
		return nil
	}
	return stateMsgAttr.S
}
