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
	if registrationData.Equals(oldTask.registrationData()) {
		return internalTask.RegistrationResultNotChanged, nil
	}
	return internalTask.RegistrationResultChanged, nil
}

func (registerer *TaskRegisterer) putTask(registrationData internalTask.RegistrationData) (*dynamodb.PutItemOutput, error) {
	currentDate := registerer.currentDateGetter.GetCurrentDate()
	taskToRegister := newTask(internalTask.Task{
		RegistrationData: registrationData,
		State:            internalTask.StateCreated,
	}, calculateTTL(currentDate, registerer.tasksStoringDuration))

	condExpr := "(attribute_not_exists(#process_id) and attribute_not_exists(#processing_state)) or (#state = :stateCreated and #expiration_time > :currentTime)"
	return registerer.dynamoAPI.PutItem(&dynamodb.PutItemInput{
		ConditionExpression: &condExpr,
		ExpressionAttributeNames: map[string]*string{
			"#process_id":       aws.String("process_id"),
			"#processing_state": aws.String("processing_state"),
			"#state":            aws.String("state"),
			"#expiration_time":  aws.String("expiration_time"),
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
