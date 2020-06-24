package main

import (
	"github.com/artii15/termination-detector/internal/api/handlers"
	"github.com/artii15/termination-detector/internal/dynamo"
	"github.com/artii15/termination-detector/pkg/dates"
	"github.com/artii15/termination-detector/pkg/env"
	"github.com/artii15/termination-detector/pkg/http"
	lambdaHandlers "github.com/artii15/termination-detector/pkg/lambda"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

const (
	tasksTableNameEnvVar       = "TASKS_TABLE_NAME"
	tasksStoringDurationEnvVar = "TASKS_STORING_DURATION"
)

func main() {
	tasksTableName := env.MustRead(tasksTableNameEnvVar)
	tasksStoringDuration := dates.MustParseDuration(env.MustRead(tasksStoringDurationEnvVar))

	awsSess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	dynamoAPI := dynamodb.New(awsSess)

	currentDateGetter := dates.NewCurrentDateGetter()
	taskRegisterer := dynamo.NewTaskRegisterer(dynamoAPI, tasksTableName, currentDateGetter, tasksStoringDuration)
	putTaskRequestHandler := handlers.NewPutTaskRequestHandler(taskRegisterer)
	taskCompleter := dynamo.NewTaskCompleter(dynamoAPI, tasksTableName, currentDateGetter)
	putTaskCompletionRequestHandler := handlers.NewPutTaskCompletionRequestHandler(taskCompleter)
	processGetter := dynamo.NewProcessGetter(dynamoAPI, tasksTableName, currentDateGetter)
	getProcessRequestHandler := handlers.NewGetProcessRequestHandler(processGetter)
	router := http.NewRouter(map[http.ResourcePath]map[http.Method]http.RequestHandler{
		http.ResourcePathTask: {
			http.MethodPut: putTaskRequestHandler,
		},
		http.ResourcePathTaskCompletion: {
			http.MethodPut: putTaskCompletionRequestHandler,
		},
		http.ResourcePathProcess: {
			http.MethodGet: getProcessRequestHandler,
		},
	})
	handler := lambdaHandlers.NewAPIGatewayEventHandler(router)
	lambda.Start(handler.Handle)
}
