package main

import (
	"github.com/artii15/termination-detector/internal/api"
	"github.com/artii15/termination-detector/internal/api/handlers"
	"github.com/artii15/termination-detector/internal/dates"
	"github.com/artii15/termination-detector/internal/dynamo"
	"github.com/artii15/termination-detector/internal/env"
	"github.com/artii15/termination-detector/pkg/http"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

const (
	tasksTableNameEnvVar       = "TASKS_TABLE_NAME"
	tasksStoringDurationEnvVar = "TASKS_STORING_DURATION"
)

type apiGatewayEventHandler struct {
	router *api.Router
}

func (handler *apiGatewayEventHandler) handle(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	routerRequest := http.Request{
		Method:         http.Method(request.HTTPMethod),
		ResourcePath:   http.ResourcePath(request.Resource),
		Body:           request.Body,
		PathParameters: readPathParameters(request.PathParameters),
	}
	response := handler.router.Route(routerRequest)
	return events.APIGatewayProxyResponse{
		StatusCode: response.StatusCode,
		Headers:    response.Headers,
		Body:       response.Body,
	}, nil
}

func readPathParameters(parameters map[string]string) map[http.PathParameter]string {
	pathParameters := make(map[http.PathParameter]string)
	for parameterName, parameterValue := range parameters {
		pathParameters[http.PathParameter(parameterName)] = parameterValue
	}
	return pathParameters
}

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
	router := api.NewRouter(map[http.ResourcePath]map[http.Method]api.RequestHandler{
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
	handler := apiGatewayEventHandler{router: router}
	lambda.Start(handler.handle)
}
