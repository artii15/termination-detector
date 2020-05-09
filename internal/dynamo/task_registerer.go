package dynamo

import (
	"time"

	"github.com/aws/aws-sdk-go/aws/awserr"

	"github.com/aws/aws-sdk-go/aws"

	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
	internalTask "github.com/nordcloud/termination-detector/internal/task"
)

type currentDateGetter interface {
	GetCurrentDate() time.Time
}

type TaskRegisterer struct {
	dynamoAPI            dynamodbiface.DynamoDBAPI
	tasksTableName       string
	currentDateGetter    currentDateGetter
	tasksStoringDuration time.Duration
}

func NewTaskRegisterer(dynamoAPI dynamodbiface.DynamoDBAPI, tasksTableName string,
	currentDateGetter currentDateGetter, tasksStoringDuration time.Duration) *TaskRegisterer {
	return &TaskRegisterer{
		dynamoAPI:            dynamoAPI,
		tasksTableName:       tasksTableName,
		currentDateGetter:    currentDateGetter,
		tasksStoringDuration: tasksStoringDuration,
	}
}

func (registerer *TaskRegisterer) Register(registrationData internalTask.RegistrationData) (internalTask.RegistrationResult, error) {
	putTaskOutput, err := registerer.putTask(registrationData)
	if err != nil {
		if awsErr, isAWSErr := err.(awserr.Error); isAWSErr && awsErr.Code() == dynamodb.ErrCodeConditionalCheckFailedException {
			return internalTask.RegistrationResultErrorTerminalState, nil
		}
		return "", err
	}
	if len(putTaskOutput.Attributes) == 0 {
		return internalTask.RegistrationResultCreated, nil
	}
	oldTask := readDynamoTask(putTaskOutput.Attributes)
	if registrationData.ID.Equals(oldTask.ID()) && registrationData.ExpirationTime.Equal(oldTask.ExpirationTime) {
		return internalTask.RegistrationResultNotChanged, nil
	}
	return internalTask.RegistrationResultChanged, nil
}

func (registerer *TaskRegisterer) putTask(registrationData internalTask.RegistrationData) (*dynamodb.PutItemOutput, error) {
	currentDate := registerer.currentDateGetter.GetCurrentDate()
	taskToRegister := newTask(internalTask.Task{
		ID:             registrationData.ID,
		ExpirationTime: registrationData.ExpirationTime,
		State:          internalTask.StateCreated,
	}, calculateTTL(currentDate, registerer.tasksStoringDuration))

	condExpr := `(attribute_not_exists(#processID) and attribute_not_exists(#taskID)) or 
		(#state = :stateCreated and #expirationTime > :currentTime)`
	return registerer.dynamoAPI.PutItem(&dynamodb.PutItemInput{
		ConditionExpression: &condExpr,
		ExpressionAttributeNames: map[string]*string{
			"#processID":      aws.String("process_id"),
			"#taskID":         aws.String("task_id"),
			"#state":          aws.String("state"),
			"#expirationTime": aws.String("expiration_time"),
		},
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":stateCreated": {S: aws.String(string(internalTask.StateCreated))},
			":currentTime":  {S: aws.String(currentDate.Format(time.RFC3339))},
		},
		Item:         taskToRegister.dynamoItem(),
		TableName:    &registerer.tasksTableName,
		ReturnValues: aws.String(dynamodb.ReturnValueAllOld),
	})
}

func calculateTTL(currentDate time.Time, itemStoringDuration time.Duration) int64 {
	return currentDate.Add(itemStoringDuration).Unix()
}
