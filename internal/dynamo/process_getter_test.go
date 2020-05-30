package dynamo_test

import (
	"errors"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/nordcloud/termination-detector/internal/task"

	"github.com/nordcloud/termination-detector/internal/process"

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

func TestProcessGetter_Get_CompletedProcess(t *testing.T) {
	procGetterAndMocks := newProcessGetterWithMocks()
	procID := "1"
	checkIfProcExistsQueryInput := dynamo.BuildCheckIfProcessExistsQueryInput(tasksTableName, procID)
	procGetterAndMocks.dynamoAPI.On("Query", checkIfProcExistsQueryInput).Return(&dynamodb.QueryOutput{
		Items: []map[string]*dynamodb.AttributeValue{
			{dynamo.ProcessIDAttrName: {S: &procID}},
		},
	}, nil)
	getProcessQueryInput := dynamo.BuildGetProcessQueryInput(tasksTableName, procID)
	procGetterAndMocks.dynamoAPI.On("Query", getProcessQueryInput).Return(&dynamodb.QueryOutput{
		Items: nil,
	}, nil)

	proc, err := procGetterAndMocks.processGetter.Get(procID)
	assert.NoError(t, err)
	assert.NotNil(t, proc)
	assert.Equal(t, process.Process{
		ID:           procID,
		State:        process.StateCompleted,
		StateMessage: nil,
	}, *proc)
	procGetterAndMocks.assertExpectations(t)
}

func TestProcessGetter_Get_ErrorWhileGettingProcess(t *testing.T) {
	procGetterAndMocks := newProcessGetterWithMocks()
	procID := "1"
	checkIfProcExistsQueryInput := dynamo.BuildCheckIfProcessExistsQueryInput(tasksTableName, procID)
	procGetterAndMocks.dynamoAPI.On("Query", checkIfProcExistsQueryInput).Return(&dynamodb.QueryOutput{
		Items: []map[string]*dynamodb.AttributeValue{
			{dynamo.ProcessIDAttrName: {S: &procID}},
		},
	}, nil)
	getProcessQueryInput := dynamo.BuildGetProcessQueryInput(tasksTableName, procID)
	procGetterAndMocks.dynamoAPI.On("Query", getProcessQueryInput).
		Return((*dynamodb.QueryOutput)(nil), errors.New("error"))

	_, err := procGetterAndMocks.processGetter.Get(procID)
	assert.Error(t, err)
	procGetterAndMocks.assertExpectations(t)
}

func TestProcessGetter_Get_AbortedProcess(t *testing.T) {
	procGetterAndMocks := newProcessGetterWithMocks()
	procID := "1"
	checkIfProcExistsQueryInput := dynamo.BuildCheckIfProcessExistsQueryInput(tasksTableName, procID)
	procGetterAndMocks.dynamoAPI.On("Query", checkIfProcExistsQueryInput).Return(&dynamodb.QueryOutput{
		Items: []map[string]*dynamodb.AttributeValue{
			{dynamo.ProcessIDAttrName: {S: &procID}},
		},
	}, nil)

	getProcessQueryInput := dynamo.BuildGetProcessQueryInput(tasksTableName, procID)
	processFailureReason := "failure"
	procGetterAndMocks.dynamoAPI.On("Query", getProcessQueryInput).Return(&dynamodb.QueryOutput{
		Items: []map[string]*dynamodb.AttributeValue{
			{
				dynamo.ProcessIDAttrName:        {S: &procID},
				dynamo.TaskStateAttrName:        {S: aws.String(string(task.StateAborted))},
				dynamo.TaskStateMessageAttrName: {S: &processFailureReason},
			},
		},
	}, nil)

	proc, err := procGetterAndMocks.processGetter.Get(procID)
	assert.NoError(t, err)
	assert.NotNil(t, proc)
	assert.Equal(t, process.Process{
		ID:           procID,
		State:        process.StateError,
		StateMessage: &processFailureReason,
	}, *proc)
	procGetterAndMocks.assertExpectations(t)
}

func TestProcessGetter_Get_ProcessInInvalidState(t *testing.T) {
	procGetterAndMocks := newProcessGetterWithMocks()
	procID := "1"
	checkIfProcExistsQueryInput := dynamo.BuildCheckIfProcessExistsQueryInput(tasksTableName, procID)
	procGetterAndMocks.dynamoAPI.On("Query", checkIfProcExistsQueryInput).Return(&dynamodb.QueryOutput{
		Items: []map[string]*dynamodb.AttributeValue{
			{dynamo.ProcessIDAttrName: {S: &procID}},
		},
	}, nil)

	getProcessQueryInput := dynamo.BuildGetProcessQueryInput(tasksTableName, procID)
	badStateEnterTime := time.Now().UTC().Format(time.RFC3339)
	procGetterAndMocks.dynamoAPI.On("Query", getProcessQueryInput).Return(&dynamodb.QueryOutput{
		Items: []map[string]*dynamodb.AttributeValue{
			{
				dynamo.ProcessIDAttrName:             {S: &procID},
				dynamo.TaskStateAttrName:             {S: aws.String("unknown")},
				dynamo.TaskBadStateEnterTimeAttrName: {S: &badStateEnterTime},
			},
		},
	}, nil)

	_, err := procGetterAndMocks.processGetter.Get(procID)
	assert.Error(t, err)
	procGetterAndMocks.assertExpectations(t)
}

func TestProcessGetter_Get_UndefinedBadStateEnterTime(t *testing.T) {
	procGetterAndMocks := newProcessGetterWithMocks()
	procID := "1"
	checkIfProcExistsQueryInput := dynamo.BuildCheckIfProcessExistsQueryInput(tasksTableName, procID)
	procGetterAndMocks.dynamoAPI.On("Query", checkIfProcExistsQueryInput).Return(&dynamodb.QueryOutput{
		Items: []map[string]*dynamodb.AttributeValue{
			{dynamo.ProcessIDAttrName: {S: &procID}},
		},
	}, nil)

	getProcessQueryInput := dynamo.BuildGetProcessQueryInput(tasksTableName, procID)
	procGetterAndMocks.dynamoAPI.On("Query", getProcessQueryInput).Return(&dynamodb.QueryOutput{
		Items: []map[string]*dynamodb.AttributeValue{
			{
				dynamo.ProcessIDAttrName: {S: &procID},
				dynamo.TaskStateAttrName: {S: aws.String(string(task.StateCreated))},
			},
		},
	}, nil)

	_, err := procGetterAndMocks.processGetter.Get(procID)
	assert.Error(t, err)
	procGetterAndMocks.assertExpectations(t)
}

func TestProcessGetter_Get_ProcessTimedOut(t *testing.T) {
	procGetterAndMocks := newProcessGetterWithMocks()
	procID := "1"
	checkIfProcExistsQueryInput := dynamo.BuildCheckIfProcessExistsQueryInput(tasksTableName, procID)
	procGetterAndMocks.dynamoAPI.On("Query", checkIfProcExistsQueryInput).Return(&dynamodb.QueryOutput{
		Items: []map[string]*dynamodb.AttributeValue{
			{dynamo.ProcessIDAttrName: {S: &procID}},
		},
	}, nil)

	currentTime := time.Now().UTC()
	procGetterAndMocks.currentDateGetter.On("GetCurrentDate").Return(currentTime)
	taskBadStateEnterTimeString := currentTime.Add(-time.Hour).Format(time.RFC3339)
	getProcessQueryInput := dynamo.BuildGetProcessQueryInput(tasksTableName, procID)
	procGetterAndMocks.dynamoAPI.On("Query", getProcessQueryInput).Return(&dynamodb.QueryOutput{
		Items: []map[string]*dynamodb.AttributeValue{
			{
				dynamo.ProcessIDAttrName:             {S: &procID},
				dynamo.TaskStateAttrName:             {S: aws.String(string(task.StateCreated))},
				dynamo.TaskBadStateEnterTimeAttrName: {S: &taskBadStateEnterTimeString},
			},
		},
	}, nil)

	proc, err := procGetterAndMocks.processGetter.Get(procID)
	assert.NoError(t, err)
	assert.NotNil(t, proc)
	assert.Equal(t, &process.Process{
		ID:           procID,
		State:        process.StateError,
		StateMessage: aws.String(process.TimedOutErrorMessage),
	}, proc)
	procGetterAndMocks.assertExpectations(t)
}

func TestProcessGetter_Get_ProcessIsWaiting(t *testing.T) {
	procGetterAndMocks := newProcessGetterWithMocks()
	procID := "1"
	checkIfProcExistsQueryInput := dynamo.BuildCheckIfProcessExistsQueryInput(tasksTableName, procID)
	procGetterAndMocks.dynamoAPI.On("Query", checkIfProcExistsQueryInput).Return(&dynamodb.QueryOutput{
		Items: []map[string]*dynamodb.AttributeValue{
			{dynamo.ProcessIDAttrName: {S: &procID}},
		},
	}, nil)

	currentTime := time.Now().UTC()
	procGetterAndMocks.currentDateGetter.On("GetCurrentDate").Return(currentTime)
	taskBadStateEnterTimeString := currentTime.Add(time.Hour).Format(time.RFC3339)
	getProcessQueryInput := dynamo.BuildGetProcessQueryInput(tasksTableName, procID)
	procGetterAndMocks.dynamoAPI.On("Query", getProcessQueryInput).Return(&dynamodb.QueryOutput{
		Items: []map[string]*dynamodb.AttributeValue{
			{
				dynamo.ProcessIDAttrName:             {S: &procID},
				dynamo.TaskStateAttrName:             {S: aws.String(string(task.StateCreated))},
				dynamo.TaskBadStateEnterTimeAttrName: {S: &taskBadStateEnterTimeString},
			},
		},
	}, nil)

	proc, err := procGetterAndMocks.processGetter.Get(procID)
	assert.NoError(t, err)
	assert.NotNil(t, proc)
	assert.Equal(t, &process.Process{
		ID:    procID,
		State: process.StateCreated,
	}, proc)
	procGetterAndMocks.assertExpectations(t)
}

func TestProcessGetter_Get_UndefinedProcessState(t *testing.T) {
	procGetterAndMocks := newProcessGetterWithMocks()
	procID := "1"
	checkIfProcExistsQueryInput := dynamo.BuildCheckIfProcessExistsQueryInput(tasksTableName, procID)
	procGetterAndMocks.dynamoAPI.On("Query", checkIfProcExistsQueryInput).Return(&dynamodb.QueryOutput{
		Items: []map[string]*dynamodb.AttributeValue{
			{dynamo.ProcessIDAttrName: {S: &procID}},
		},
	}, nil)

	getProcessQueryInput := dynamo.BuildGetProcessQueryInput(tasksTableName, procID)
	procGetterAndMocks.dynamoAPI.On("Query", getProcessQueryInput).Return(&dynamodb.QueryOutput{
		Items: []map[string]*dynamodb.AttributeValue{
			{
				dynamo.ProcessIDAttrName: {S: &procID},
			},
		},
	}, nil)

	_, err := procGetterAndMocks.processGetter.Get(procID)
	assert.Error(t, err)
	procGetterAndMocks.assertExpectations(t)
}
