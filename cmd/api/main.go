package main

import (
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/nordcloud/termination-detector/internal/api"
	"github.com/nordcloud/termination-detector/internal/api/handlers"
	"github.com/nordcloud/termination-detector/internal/dates"
	"github.com/nordcloud/termination-detector/internal/dynamo"
	"github.com/nordcloud/termination-detector/internal/env"
)

const (
	tasksTableNameEnvVar       = "TASKS_TABLE_NAME"
	tasksStoringDurationEnvVar = "TASKS_STORING_DURATION"
)

type apiGatewayEventHandler struct {
	router *api.Router
}

func (handler *apiGatewayEventHandler) handle(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	routerRequest := api.Request{
		HTTPMethod:     api.HTTPMethod(request.HTTPMethod),
		ResourcePath:   api.ResourcePath(request.Resource),
		Body:           request.Body,
		PathParameters: request.PathParameters,
	}
	response := handler.router.Route(routerRequest)
	return events.APIGatewayProxyResponse{
		StatusCode: response.StatusCode,
		Headers:    response.Headers,
		Body:       response.Body,
	}, nil
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
	router := api.NewRouter(map[api.ResourcePath]map[api.HTTPMethod]api.RequestHandler{
		api.ResourcePathTask: {
			api.HTTPMethodPut: putTaskRequestHandler,
		},
	})
	handler := apiGatewayEventHandler{router: router}
	lambda.Start(handler.handle)
}
