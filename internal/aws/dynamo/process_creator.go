package dynamo

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
	"github.com/nordcloud/termination-detector/internal/process"
	"github.com/pkg/errors"
)

type ProcessCreator struct {
	dynamoAPI          dynamodbiface.DynamoDBAPI
	processesTableName string
}

func NewProcessCreator(dynamoAPI dynamodbiface.DynamoDBAPI, processesTableName string) *ProcessCreator {
	return &ProcessCreator{
		dynamoAPI:          dynamoAPI,
		processesTableName: processesTableName,
	}
}

func (creator *ProcessCreator) Create(proc process.Process) (process.CreationStatus, error) {
	dynamoItem, err := newStoredProcess(proc).dynamoItem()
	if err != nil {
		return "", err
	}

	putItemInput := &dynamodb.PutItemInput{
		ConditionExpression: aws.String("attribute_not_exists(#id) or #state = :stateCreated"),
		ExpressionAttributeNames: map[string]*string{
			"#id":    aws.String("id"),
			"#state": aws.String("state"),
		},
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":stateCreated": {S: aws.String(string(process.StateCreated))},
		},
		Item:         dynamoItem,
		ReturnValues: aws.String(dynamodb.ReturnValueAllOld),
		TableName:    &creator.processesTableName,
	}
	return determineProcessCreatingStatus(creator.dynamoAPI.PutItem(putItemInput))
}

func determineProcessCreatingStatus(putItemResult *dynamodb.PutItemOutput, err error) (process.CreationStatus, error) {
	if err != nil {
		if awsErr, isAWSErr := err.(awserr.Error); isAWSErr && awsErr.Code() == dynamodb.ErrCodeConditionalCheckFailedException {
			return process.CreationStatusConflict, nil
		}
		return "", errors.Wrap(err, "failed to put process into DynamoDB table")
	}

	if len(putItemResult.Attributes) == 0 {
		return process.CreationStatusNew, nil
	}
	return process.CreationStatusOverridden, nil
}
