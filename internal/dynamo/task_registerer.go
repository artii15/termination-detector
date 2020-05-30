package dynamo

import (
	"fmt"
	"strconv"
	"time"

	"github.com/nordcloud/termination-detector/pkg/task"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
)

const (
	decimalBase                        = 10
	taskTTLValuePlaceholder            = ":ttl"
	taskExpirationTimeValuePlaceholder = ":expirationTime"
)

var (
	registerTaskConditionExpr = fmt.Sprintf("attribute_not_exists(%s) and attribute_not_exists(%s)",
		processIDAttrAlias, taskIDAttrAlias)
	registerTaskUpdateExpr = fmt.Sprintf(`SET %s = %s, %s = %s, %s = %s, %s = %s`,
		taskExpirationTimeAttrAlias, taskExpirationTimeValuePlaceholder, taskStateAttrAlias, taskStateCreatedValuePlaceholder,
		taskTTLAttrAlias, taskTTLValuePlaceholder, taskBadStateEnterTimeAttrAlias, taskBadStateEnterTimeValuePlaceholder)
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

func (registerer *TaskRegisterer) Register(registrationData task.RegistrationData) (task.RegistrationResult, error) {
	if err := registerer.saveTask(registrationData); err != nil {
		if awsErr, isAWSErr := err.(awserr.Error); isAWSErr && awsErr.Code() == dynamodb.ErrCodeConditionalCheckFailedException {
			return task.RegistrationResultAlreadyRegistered, nil
		}
		return "", err
	}
	return task.RegistrationResultCreated, nil
}

func (registerer *TaskRegisterer) saveTask(registrationData task.RegistrationData) error {
	updateItemInput := BuildRegisterTaskUpdateItemInput(registerer.tasksTableName, TaskToRegister{
		CreationTime:     registerer.currentDateGetter.GetCurrentDate(),
		StoringDuration:  registerer.tasksStoringDuration,
		RegistrationData: registrationData,
	})
	_, err := registerer.dynamoAPI.UpdateItem(updateItemInput)
	return err
}

type TaskToRegister struct {
	CreationTime     time.Time
	StoringDuration  time.Duration
	RegistrationData task.RegistrationData
}

func BuildRegisterTaskUpdateItemInput(tableName string, taskToRegister TaskToRegister) *dynamodb.UpdateItemInput {
	ttl := taskToRegister.CreationTime.Add(taskToRegister.StoringDuration).UTC().Unix()
	ttlString := strconv.FormatInt(ttl, decimalBase)
	expirationTimeString := taskToRegister.RegistrationData.ExpirationTime.Format(time.RFC3339)
	return &dynamodb.UpdateItemInput{
		ConditionExpression: &registerTaskConditionExpr,
		ExpressionAttributeNames: map[string]*string{
			processIDAttrAlias:             aws.String(ProcessIDAttrName),
			taskIDAttrAlias:                aws.String(taskIDAttrName),
			taskStateAttrAlias:             aws.String(TaskStateAttrName),
			taskExpirationTimeAttrAlias:    aws.String(taskExpirationTimeAttrName),
			taskTTLAttrAlias:               aws.String(taskTTLAttributeName),
			taskBadStateEnterTimeAttrAlias: aws.String(TaskBadStateEnterTimeAttrName),
		},
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			taskStateCreatedValuePlaceholder:      {S: aws.String(string(task.StateCreated))},
			taskTTLValuePlaceholder:               {N: &ttlString},
			taskExpirationTimeValuePlaceholder:    {S: &expirationTimeString},
			taskBadStateEnterTimeValuePlaceholder: {S: &expirationTimeString},
		},
		UpdateExpression: &registerTaskUpdateExpr,
		TableName:        &tableName,
		Key: map[string]*dynamodb.AttributeValue{
			ProcessIDAttrName: {S: &taskToRegister.RegistrationData.ID.ProcessID},
			taskIDAttrName:    {S: &taskToRegister.RegistrationData.ID.TaskID},
		},
	}
}
