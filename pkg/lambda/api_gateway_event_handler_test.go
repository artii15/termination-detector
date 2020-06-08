package lambda_test

import (
	"net/http"
	"testing"
	"time"

	internalHTTP "github.com/artii15/termination-detector/pkg/http"
	"github.com/artii15/termination-detector/pkg/lambda"
	"github.com/aws/aws-lambda-go/events"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type routerMock struct {
	mock.Mock
}

func (router *routerMock) Route(request internalHTTP.Request) internalHTTP.Response {
	return router.Called(request).Get(0).(internalHTTP.Response)
}

type apiGatewayEventHandlerWithMocks struct {
	handler *lambda.APIGatewayEventHandler
	router  *routerMock
}

func newAPIGatewayEventHandlerWithMocks() *apiGatewayEventHandlerWithMocks {
	router := new(routerMock)
	return &apiGatewayEventHandlerWithMocks{
		handler: lambda.NewAPIGatewayEventHandler(router),
		router:  router,
	}
}

func TestAPIGatewayEventHandler_Handle(t *testing.T) {
	handlerAndMocks := newAPIGatewayEventHandlerWithMocks()

	method := internalHTTP.MethodPut
	resource := internalHTTP.ResourcePathTask
	task := internalHTTP.Task{ExpirationTime: time.Now()}
	procID := "1"
	taskID := "2"
	requestBody := task.JSON()

	responseFromRouter := internalHTTP.Response{
		StatusCode: http.StatusOK,
		Body:       "{\"field\": \"value\"}",
		Headers: map[string]string{
			internalHTTP.ContentTypeHeaderName: internalHTTP.ContentTypeApplicationJSON,
		},
	}
	handlerAndMocks.router.On("Route", internalHTTP.Request{
		Method:       method,
		ResourcePath: resource,
		Body:         requestBody,
		PathParameters: map[internalHTTP.PathParameter]string{
			internalHTTP.PathParameterProcessID: procID,
			internalHTTP.PathParameterTaskID:    taskID,
		},
	}).Return(responseFromRouter)

	response, err := handlerAndMocks.handler.Handle(events.APIGatewayProxyRequest{
		Resource:   string(resource),
		HTTPMethod: string(method),
		PathParameters: map[string]string{
			string(internalHTTP.PathParameterProcessID): procID,
			string(internalHTTP.PathParameterTaskID):    taskID,
		},
		Body: requestBody,
	})
	assert.NoError(t, err)
	assert.Equal(t, events.APIGatewayProxyResponse{
		StatusCode: responseFromRouter.StatusCode,
		Headers: map[string]string{
			internalHTTP.ContentTypeHeaderName: internalHTTP.ContentTypeApplicationJSON,
		},
		Body: responseFromRouter.Body,
	}, response)
}
