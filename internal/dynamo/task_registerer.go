package dynamo

import (
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws/awserr"

	"github.com/aws/aws-sdk-go/aws"

	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
	internalTask "github.com/nordcloud/termination-detector/internal/task"
)

const decimalBase = 10

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
	if err := registerer.putTask(registrationData); err != nil {
		if awsErr, isAWSErr := err.(awserr.Error); isAWSErr && awsErr.Code() == dynamodb.ErrCodeConditionalCheckFailedException {
			return internalTask.RegistrationResultAlreadyRegistered, nil
		}
		return "", err
	}
	return internalTask.RegistrationResultCreated, nil
}

func (registerer *TaskRegisterer) putTask(registrationData internalTask.RegistrationData) error {
	currentDate := registerer.currentDateGetter.GetCurrentDate()
	condExpr := `attribute_not_exists(#processID) and attribute_not_exists(#taskID)`
	updateExpr := `SET #expirationTime = :expirationTime, #state = :stateCreated, #ttl = :ttl, #badStateEnterTime = :badStateEnterTime`
	ttl := currentDate.Add(registerer.tasksStoringDuration).UTC().Unix()
	ttlString := strconv.FormatInt(ttl, decimalBase)
	expirationTimeString := registrationData.ExpirationTime.Format(time.RFC3339)
	_, err := registerer.dynamoAPI.UpdateItem(&dynamodb.UpdateItemInput{
		ConditionExpression: &condExpr,
		ExpressionAttributeNames: map[string]*string{
			"#processID":         aws.String("process_id"),
			"#taskID":            aws.String("task_id"),
			"#state":             aws.String("state"),
			"#expirationTime":    aws.String("expiration_time"),
			"#ttl":               aws.String("ttl"),
			"#badStateEnterTime": aws.String("bad_state_enter_time"),
		},
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":stateCreated":      {S: aws.String(string(internalTask.StateCreated))},
			":ttl":               {N: &ttlString},
			":expirationTime":    {S: &expirationTimeString},
			":badStateEnterTime": {S: &expirationTimeString},
		},
		UpdateExpression: &updateExpr,
		TableName:        &registerer.tasksTableName,
		Key: map[string]*dynamodb.AttributeValue{
			"process_id": {S: &registrationData.ID.ProcessID},
			"task_id":    {S: &registrationData.ID.TaskID},
		},
	})
	return err
}
