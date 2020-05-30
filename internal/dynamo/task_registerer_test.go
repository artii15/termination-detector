package dynamo_test

import (
	"errors"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws/awserr"

	"github.com/aws/aws-sdk-go/service/dynamodb"

	"github.com/nordcloud/termination-detector/internal/dynamo"
	"github.com/nordcloud/termination-detector/internal/task"
	"github.com/stretchr/testify/assert"
)

type taskRegistererWithMocks struct {
	dynamoAPI            *dynamoAPIMock
	currentDateGetter    *currentDateGetterMock
	tasksStoringDuration time.Duration
	registerer           *dynamo.TaskRegisterer
}

func (registererAndMocks *taskRegistererWithMocks) assertExpectations(t *testing.T) {
	registererAndMocks.dynamoAPI.AssertExpectations(t)
	registererAndMocks.currentDateGetter.AssertExpectations(t)
}

func newTaskRegistererWithMocks() *taskRegistererWithMocks {
	dynamoAPI := new(dynamoAPIMock)
	currentDateGetter := new(currentDateGetterMock)
	tasksStoringDuration := time.Hour * 24 * 7
	return &taskRegistererWithMocks{
		dynamoAPI:            dynamoAPI,
		tasksStoringDuration: tasksStoringDuration,
		currentDateGetter:    currentDateGetter,
		registerer:           dynamo.NewTaskRegisterer(dynamoAPI, tasksTableName, currentDateGetter, tasksStoringDuration),
	}
}

func TestTaskRegisterer_Register(t *testing.T) {
	registererAndMocks := newTaskRegistererWithMocks()
	currentDate := time.Now().UTC()
	registrationData := task.RegistrationData{
		ID: task.ID{
			ProcessID: "2",
			TaskID:    "1",
		},
		ExpirationTime: currentDate.Add(time.Hour),
	}
	registererAndMocks.currentDateGetter.On("GetCurrentDate").Return(currentDate)
	taskToRegister := dynamo.TaskToRegister{
		CreationTime:     currentDate,
		StoringDuration:  registererAndMocks.tasksStoringDuration,
		RegistrationData: registrationData,
	}
	updateItemInput := dynamo.BuildRegisterTaskUpdateItemInput(tasksTableName, taskToRegister)
	registererAndMocks.dynamoAPI.On("UpdateItem", updateItemInput).Return(&dynamodb.UpdateItemOutput{}, nil)

	registrationResult, err := registererAndMocks.registerer.Register(registrationData)
	assert.NoError(t, err)
	assert.Equal(t, task.RegistrationResultCreated, registrationResult)
	registererAndMocks.assertExpectations(t)
}

func TestTaskRegisterer_Register_TaskAlreadyExists(t *testing.T) {
	registererAndMocks := newTaskRegistererWithMocks()
	currentDate := time.Now().UTC()
	registrationData := task.RegistrationData{
		ID: task.ID{
			ProcessID: "2",
			TaskID:    "1",
		},
		ExpirationTime: currentDate.Add(time.Hour),
	}
	registererAndMocks.currentDateGetter.On("GetCurrentDate").Return(currentDate)
	taskToRegister := dynamo.TaskToRegister{
		CreationTime:     currentDate,
		StoringDuration:  registererAndMocks.tasksStoringDuration,
		RegistrationData: registrationData,
	}
	updateItemInput := dynamo.BuildRegisterTaskUpdateItemInput(tasksTableName, taskToRegister)
	errToReturn := awserr.New(dynamodb.ErrCodeConditionalCheckFailedException, "", nil)
	registererAndMocks.dynamoAPI.On("UpdateItem", updateItemInput).Return((*dynamodb.UpdateItemOutput)(nil), errToReturn)

	registrationResult, err := registererAndMocks.registerer.Register(registrationData)
	assert.NoError(t, err)
	assert.Equal(t, task.RegistrationResultAlreadyRegistered, registrationResult)
	registererAndMocks.assertExpectations(t)
}

func TestTaskRegisterer_Register_UnexpectedError(t *testing.T) {
	registererAndMocks := newTaskRegistererWithMocks()
	currentDate := time.Now().UTC()
	registrationData := task.RegistrationData{
		ID: task.ID{
			ProcessID: "2",
			TaskID:    "1",
		},
		ExpirationTime: currentDate.Add(time.Hour),
	}
	registererAndMocks.currentDateGetter.On("GetCurrentDate").Return(currentDate)
	taskToRegister := dynamo.TaskToRegister{
		CreationTime:     currentDate,
		StoringDuration:  registererAndMocks.tasksStoringDuration,
		RegistrationData: registrationData,
	}
	updateItemInput := dynamo.BuildRegisterTaskUpdateItemInput(tasksTableName, taskToRegister)
	errToReturn := errors.New("error")
	registererAndMocks.dynamoAPI.On("UpdateItem", updateItemInput).Return((*dynamodb.UpdateItemOutput)(nil), errToReturn)

	_, err := registererAndMocks.registerer.Register(registrationData)
	assert.Error(t, err)
	registererAndMocks.assertExpectations(t)
}
