package api_test

import (
	"fmt"
	"os"
	"testing"
	"time"

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
