package dynamo_test

import (
	"time"

	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
	"github.com/stretchr/testify/mock"
)

const tasksTableName = "tasksTable"

type dynamoAPIMock struct {
	mock.Mock
	dynamodbiface.DynamoDBAPI
}

func (api *dynamoAPIMock) Query(input *dynamodb.QueryInput) (*dynamodb.QueryOutput, error) {
	args := api.Called(input)
	if args.Get(0) == 0 {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dynamodb.QueryOutput), args.Error(1)
}

func (api *dynamoAPIMock) UpdateItem(input *dynamodb.UpdateItemInput) (*dynamodb.UpdateItemOutput, error) {
	args := api.Called(input)
	if args.Get(0) == 0 {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dynamodb.UpdateItemOutput), args.Error(1)
}

type currentDateGetterMock struct {
	mock.Mock
}

func (getter *currentDateGetterMock) GetCurrentDate() time.Time {
	return getter.Called().Get(0).(time.Time)
}
