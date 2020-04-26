package api

import (
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	log "github.com/sirupsen/logrus"
)

type ResourcePath string
type HTTPMethod string
type RequestsHandlersMap map[ResourcePath]map[HTTPMethod]RequestHandler

const (
	ResourcePathTask ResourcePath = "/processes/{process-id}/tasks/{task-id}"

	HTTPMethodGet HTTPMethod = http.MethodGet
)

type RequestHandler interface {
	HandleRequest(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error)
}

type Router struct {
	requestsHandlers RequestsHandlersMap
}

func NewRouter(requestsHandlers RequestsHandlersMap) *Router {
	return &Router{
		requestsHandlers: requestsHandlers,
	}
}

func (router *Router) Route(request events.APIGatewayProxyRequest) events.APIGatewayProxyResponse {
	methodsHandlers, handlersForResourceExist := router.requestsHandlers[ResourcePath(request.Resource)]
	if !handlersForResourceExist {
		return createDefaultResponseWithStatus(http.StatusNotFound)
	}

	requestHandler, handlerExists := methodsHandlers[HTTPMethod(request.HTTPMethod)]
	if !handlerExists {
		return createDefaultResponseWithStatus(http.StatusNotFound)
	}

	response, err := requestHandler.HandleRequest(request)
	if err != nil {
		log.WithError(err).WithField("request", request).Error("failed to handle request")
		return createDefaultResponseWithStatus(http.StatusInternalServerError)
	}
	return response
}

func createDefaultResponseWithStatus(statusCode int) events.APIGatewayProxyResponse {
	return events.APIGatewayProxyResponse{
		StatusCode: statusCode,
		Body:       http.StatusText(statusCode),
	}
}
