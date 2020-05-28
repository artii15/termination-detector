package dynamo

import (
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"

	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
	internalTask "github.com/nordcloud/termination-detector/internal/task"
)

const taskBadStateEnterTimeZeroValue = "0"

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

func (completer *TaskCompleter) Complete(request internalTask.CompleteRequest) (internalTask.CompletingResult, error) {
	if err := completer.updateTask(request); err != nil {
		if awsErr, isAWSErr := err.(awserr.Error); isAWSErr && awsErr.Code() == dynamodb.ErrCodeConditionalCheckFailedException {
			return internalTask.CompletingResultConflict, nil
		}
		return "", err
	}
	return internalTask.CompletingResultCompleted, nil
}

func (completer *TaskCompleter) updateTask(request internalTask.CompleteRequest) error {
	currentTime := completer.currentDateGetter.GetCurrentDate()
	conditionExpr := `attribute_exists(#processID) and attribute_exists(#taskID) and 
		#expirationTime > :currentTime and #state = :stateCreated`
	expressionAttributeValues := map[string]*dynamodb.AttributeValue{
		":currentTime":       {S: aws.String(currentTime.Format(time.RFC3339))},
		":stateCreated":      {S: aws.String(string(internalTask.StateCreated))},
		":newState":          {S: aws.String(string(request.State))},
		":newStateMessage":   {NULL: aws.Bool(true)},
		":badStateEnterTime": {S: aws.String(taskBadStateEnterTimeZeroValue)},
	}
	if request.Message != nil {
		expressionAttributeValues[":newStateMessage"] = &dynamodb.AttributeValue{S: request.Message}
	}
	if request.State == internalTask.StateAborted {
		expressionAttributeValues[":badStateEnterTime"] = &dynamodb.AttributeValue{S: aws.String(currentTime.Format(time.RFC3339))}
	}

	_, err := completer.dynamoAPI.UpdateItem(&dynamodb.UpdateItemInput{
		ConditionExpression: &conditionExpr,
		ExpressionAttributeNames: map[string]*string{
			"#processID":         aws.String("process_id"),
			"#taskID":            aws.String("task_id"),
			"#expirationTime":    aws.String("expiration_time"),
			"#state":             aws.String("state"),
			"#stateMessage":      aws.String("state_message"),
			"#badStateEnterTime": aws.String("bad_state_enter_time"),
		},
		ExpressionAttributeValues: expressionAttributeValues,
		Key: map[string]*dynamodb.AttributeValue{
			"process_id": {S: &request.ProcessID},
			"task_id":    {S: &request.TaskID},
		},
		TableName:        &completer.tasksTableName,
		UpdateExpression: aws.String("SET #state = :newState, #stateMessage = :newStateMessage, #badStateEnterTime = :badStateEnterTime"),
	})
	return err
}
