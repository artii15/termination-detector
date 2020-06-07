package api_test

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/nordcloud/termination-detector/pkg/process"

	"github.com/nordcloud/termination-detector/pkg/task"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/nordcloud/termination-detector/internal/dynamo"
	"github.com/nordcloud/termination-detector/pkg/sdk"
	"github.com/stretchr/testify/assert"
)

const (
	iamAuthorizedAPIURLEnvVarName = "IAM_AUTHORIZED_API_URL"
	tasksTableNameEnvVarName      = "TASKS_TABLE_NAME"

	testProcessID   = "1"
	requestsTimeout = time.Second * 30
)

type apiIntegrationTestConfig struct {
	apiURL         string
	tasksTableName string
}

func newIAMAuthorizedAPIIntegrationTestConfig() *apiIntegrationTestConfig {
	apiURL, isAPIURLDefined := os.LookupEnv(iamAuthorizedAPIURLEnvVarName)
	tasksTableName, isTasksTableNameDefined := os.LookupEnv(tasksTableNameEnvVarName)
	if !isAPIURLDefined || !isTasksTableNameDefined {
		return nil
	}
	return &apiIntegrationTestConfig{
		apiURL:         apiURL,
		tasksTableName: tasksTableName,
	}
}

func TestUsingAWSIAMAuthorizedSDK(t *testing.T) {
	apiTestConfig := newIAMAuthorizedAPIIntegrationTestConfig()
	if apiTestConfig == nil {
		t.Skip("env variables required for API integration tests are not defined")
	}
	awsSess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState:       session.SharedConfigEnable,
		AssumeRoleTokenProvider: stscreds.StdinTokenProvider,
	}))
	terminationDetectorSDK := sdk.NewAWSIAMAuthorized(requestsTimeout, apiTestConfig.apiURL, *awsSess.Config.Region, awsSess.Config.Credentials)
	defer removeTestDataFromDB(t, awsSess, apiTestConfig.tasksTableName)

	t.Run("not registered process not exists", func(t *testing.T) {
		proc, err := terminationDetectorSDK.Get(testProcessID)
		assert.NoError(t, err)
		assert.Nil(t, proc)
	})

	task1ID := "1"
	task2ID := "2"
	t.Run("each task can be registered only once", func(t *testing.T) {
		registrationResult, err := terminationDetectorSDK.Register(task.RegistrationData{
			ID: task.ID{
				ProcessID: testProcessID,
				TaskID:    task1ID,
			},
			ExpirationTime: time.Now().Add(time.Hour),
		})
		assert.NoError(t, err)
		assert.Equal(t, task.RegistrationResultCreated, registrationResult)

		registrationResult, err = terminationDetectorSDK.Register(task.RegistrationData{
			ID: task.ID{
				ProcessID: testProcessID,
				TaskID:    task2ID,
			},
			ExpirationTime: time.Now().Add(time.Hour),
		})
		assert.NoError(t, err)
		assert.Equal(t, task.RegistrationResultCreated, registrationResult)

		registrationResult, err = terminationDetectorSDK.Register(task.RegistrationData{
			ID: task.ID{
				ProcessID: testProcessID,
				TaskID:    task1ID,
			},
			ExpirationTime: time.Now().Add(time.Hour),
		})
		assert.NoError(t, err)
		assert.Equal(t, task.RegistrationResultAlreadyRegistered, registrationResult)
	})
	t.Run("process is completed only when all tasks are completed", func(t *testing.T) {
		proc, err := terminationDetectorSDK.Get(testProcessID)
		assert.NoError(t, err)
		assert.NotNil(t, proc)
		assert.Equal(t, process.StateCreated, proc.State)

		completeResult, err := terminationDetectorSDK.Complete(task.CompleteRequest{
			ID: task.ID{
				ProcessID: testProcessID,
				TaskID:    task1ID,
			},
			State: task.StateFinished,
		})
		assert.NoError(t, err)
		assert.Equal(t, task.CompletingResultCompleted, completeResult)

		proc, err = terminationDetectorSDK.Get(testProcessID)
		assert.NoError(t, err)
		assert.NotNil(t, proc)
		assert.Equal(t, process.StateCreated, proc.State)

		completeResult, err = terminationDetectorSDK.Complete(task.CompleteRequest{
			ID: task.ID{
				ProcessID: testProcessID,
				TaskID:    task2ID,
			},
			State: task.StateFinished,
		})
		assert.NoError(t, err)
		assert.Equal(t, task.CompletingResultCompleted, completeResult)

		proc, err = terminationDetectorSDK.Get(testProcessID)
		assert.NoError(t, err)
		assert.NotNil(t, proc)
		assert.Equal(t, process.StateCompleted, proc.State)
	})

	task3ID := "3"
	t.Run("process fails if at least one task fails", func(t *testing.T) {
		registrationStatus, err := terminationDetectorSDK.Register(task.RegistrationData{
			ID: task.ID{
				ProcessID: testProcessID,
				TaskID:    task3ID,
			},
			ExpirationTime: time.Now().Add(time.Hour),
		})
		assert.NoError(t, err)
		assert.Equal(t, task.RegistrationResultCreated, registrationStatus)

		failureReason := "failure"
		completeResult, err := terminationDetectorSDK.Complete(task.CompleteRequest{
			ID: task.ID{
				ProcessID: testProcessID,
				TaskID:    task3ID,
			},
			State:   task.StateAborted,
			Message: &failureReason,
		})
		assert.NoError(t, err)
		assert.Equal(t, task.CompletingResultCompleted, completeResult)

		proc, err := terminationDetectorSDK.Get(testProcessID)
		assert.NoError(t, err)
		assert.NotNil(t, proc)
		assert.Equal(t, process.StateError, proc.State)
		assert.Equal(t, &failureReason, proc.StateMessage)
	})

	task4ID := "4"
	t.Run("process fails if task times out", func(t *testing.T) {
		registrationStatus, err := terminationDetectorSDK.Register(task.RegistrationData{
			ID: task.ID{
				ProcessID: testProcessID,
				TaskID:    task4ID,
			},
			ExpirationTime: time.Now().Add(-time.Hour * 24),
		})
		assert.NoError(t, err)
		assert.Equal(t, task.RegistrationResultCreated, registrationStatus)

		proc, err := terminationDetectorSDK.Get(testProcessID)
		assert.NoError(t, err)
		assert.NotNil(t, proc)
		assert.Equal(t, process.StateError, proc.State)
		assert.NotEmpty(t, proc.StateMessage)
	})
}

func removeTestDataFromDB(t *testing.T, awsSess *session.Session, tasksTableName string) {
	dynamoAPI := dynamodb.New(awsSess)
	err := dynamoAPI.QueryPages(&dynamodb.QueryInput{
		ConsistentRead:         aws.Bool(true),
		KeyConditionExpression: aws.String(fmt.Sprintf("%s = %s", dynamo.ProcessIDAttrAlias, dynamo.ProcessIDValuePlaceholder)),
		ExpressionAttributeNames: map[string]*string{
			dynamo.ProcessIDAttrAlias: aws.String(dynamo.ProcessIDAttrName),
		},
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			dynamo.ProcessIDValuePlaceholder: {S: aws.String(testProcessID)},
		},
		TableName: &tasksTableName,
	}, func(page *dynamodb.QueryOutput, isLastPage bool) bool {
		for _, item := range page.Items {
			_, err := dynamoAPI.DeleteItem(&dynamodb.DeleteItemInput{
				Key: map[string]*dynamodb.AttributeValue{
					dynamo.ProcessIDAttrName: {S: item[dynamo.ProcessIDAttrName].S},
					dynamo.TaskIDAttrName:    {S: item[dynamo.TaskIDAttrName].S},
				},
				TableName: &tasksTableName,
			})
			if err != nil {
				t.Fatalf("failed to remove test data item from DB: %s", err.Error())
			}
		}
		return !isLastPage
	})
	if err != nil {
		t.Fatalf("failed to remove test data from DB: %s", err.Error())
	}
}
