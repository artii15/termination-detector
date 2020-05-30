package api

import (
	"net/http"

	log "github.com/sirupsen/logrus"
)

type ResourcePath string
type HTTPMethod string
type RequestsHandlersMap map[ResourcePath]map[HTTPMethod]RequestHandler

const (
	ResourcePathTask           ResourcePath = "/processes/{process_id}/tasks/{task_id}"
	ResourcePathTaskCompletion ResourcePath = "/processes/{process_id}/tasks/{task_id}/completion"
	ResourcePathProcess        ResourcePath = "/processes/{process_id}"

	HTTPMethodGet HTTPMethod = http.MethodGet
	HTTPMethodPut HTTPMethod = http.MethodPut
)

type RequestHandler interface {
	HandleRequest(request Request) (Response, error)
}

type Router struct {
	requestsHandlers RequestsHandlersMap
}

func NewRouter(requestsHandlers RequestsHandlersMap) *Router {
	return &Router{
		requestsHandlers: requestsHandlers,
	}
}

type Request struct {
	HTTPMethod     HTTPMethod
	ResourcePath   ResourcePath
	Body           string
	PathParameters map[string]string
}

type Response struct {
	StatusCode int
	Body       string
	Headers    map[string]string
}

func (router *Router) Route(request Request) Response {
	methodsHandlers, handlersForResourceExist := router.requestsHandlers[request.ResourcePath]
	if !handlersForResourceExist {
		return CreateDefaultTextResponseWithStatus(http.StatusNotFound)
	}

	requestHandler, handlerExists := methodsHandlers[request.HTTPMethod]
	if !handlerExists {
		return CreateDefaultTextResponseWithStatus(http.StatusMethodNotAllowed)
	}

	response, err := requestHandler.HandleRequest(request)
	if err != nil {
		log.WithError(err).WithField("request", request).Error("failed to handle request")
		return CreateDefaultTextResponseWithStatus(http.StatusInternalServerError)
	}
	return response
}

func CreateDefaultTextResponseWithStatus(statusCode int) Response {
	return Response{
		StatusCode: statusCode,
		Body:       http.StatusText(statusCode),
		Headers: map[string]string{
			ContentTypeHeaderName: ContentTypeTextPlain,
		},
	}
}
