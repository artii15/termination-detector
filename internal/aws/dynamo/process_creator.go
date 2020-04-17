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

func (creator *ProcessCreator) Create(proc process.Process) (process.CreationResult, error) {
	_, err := creator.dynamoAPI.PutItem(&dynamodb.PutItemInput{
		ConditionExpression: aws.String("attribute_not_exists(#id) or #state = :stateCreated"),
		ExpressionAttributeNames: map[string]*string{
			"#id":    aws.String("id"),
			"#state": aws.String("state"),
		},
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":stateCreated": {S: aws.String(string(process.StateCreated))},
		},
		Item:      newStoredProcess(proc).dynamoItem(),
		TableName: &creator.processesTableName,
	})
	if err != nil {
		if awsErr, isAWSErr := err.(awserr.Error); isAWSErr && awsErr.Code() == dynamodb.ErrCodeConditionalCheckFailedException {
			return process.CreationResult{AlreadyExistsInConflictingState: true}, nil
		}
		return process.CreationResult{}, errors.Wrap(err, "failed to put process into DynamoDB table")
	}
	return process.CreationResult{AlreadyExistsInConflictingState: false}, nil
}
