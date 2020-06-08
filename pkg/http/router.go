package http

import (
	"net/http"

	log "github.com/sirupsen/logrus"
)

type RequestsHandlersMap map[ResourcePath]map[Method]RequestHandler

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

func (router *Router) Route(request Request) Response {
	methodsHandlers, handlersForResourceExist := router.requestsHandlers[request.ResourcePath]
	if !handlersForResourceExist {
		return CreateDefaultTextResponseWithStatus(http.StatusNotFound)
	}

	requestHandler, handlerExists := methodsHandlers[request.Method]
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
