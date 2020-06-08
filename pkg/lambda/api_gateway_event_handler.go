package lambda

import (
	"github.com/artii15/termination-detector/pkg/http"
	"github.com/aws/aws-lambda-go/events"
)

type router interface {
	Route(request http.Request) http.Response
}

type APIGatewayEventHandler struct {
	router router
}

func NewAPIGatewayEventHandler(router router) *APIGatewayEventHandler {
	return &APIGatewayEventHandler{router: router}
}

func (handler *APIGatewayEventHandler) Handle(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
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
