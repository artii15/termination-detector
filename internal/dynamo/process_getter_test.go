package dynamo_test

import (
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/nordcloud/termination-detector/internal/dynamo"
	"github.com/stretchr/testify/assert"
)

type processGetterWithMocks struct {
	processGetter     *dynamo.ProcessGetter
	dynamoAPI         *dynamoAPIMock
	currentDateGetter *currentDateGetterMock
}

func (getterAndMocks *processGetterWithMocks) assertExpectations(t *testing.T) {
	getterAndMocks.dynamoAPI.AssertExpectations(t)
	getterAndMocks.currentDateGetter.AssertExpectations(t)
}

func newProcessGetterWithMocks() *processGetterWithMocks {
	dynamoAPI := new(dynamoAPIMock)
	currentDateGetter := new(currentDateGetterMock)
	return &processGetterWithMocks{
		processGetter:     dynamo.NewProcessGetter(dynamoAPI, tasksTableName, currentDateGetter),
		dynamoAPI:         dynamoAPI,
		currentDateGetter: currentDateGetter,
	}
}

func TestProcessGetter_Get_ProcessNotExists(t *testing.T) {
	procGetterAndMocks := newProcessGetterWithMocks()
	procID := "1"
	checkIfProcExistsQueryInput := dynamo.BuildCheckIfProcessExistsQueryInput(tasksTableName, procID)
	procGetterAndMocks.dynamoAPI.On("Query", checkIfProcExistsQueryInput).Return(&dynamodb.QueryOutput{
		Items: nil,
	}, nil)

	proc, err := procGetterAndMocks.processGetter.Get(procID)
	assert.NoError(t, err)
	assert.Nil(t, proc)
	procGetterAndMocks.assertExpectations(t)
}

func TestProcessGetter_Get_ErrorDuringProcSearching(t *testing.T) {
	procGetterAndMocks := newProcessGetterWithMocks()
	procID := "1"
	checkIfProcExistsQueryInput := dynamo.BuildCheckIfProcessExistsQueryInput(tasksTableName, procID)
	procGetterAndMocks.dynamoAPI.On("Query", checkIfProcExistsQueryInput).
		Return((*dynamodb.QueryOutput)(nil), errors.New("error"))

	_, err := procGetterAndMocks.processGetter.Get(procID)
	assert.Error(t, err)
	procGetterAndMocks.assertExpectations(t)
}
