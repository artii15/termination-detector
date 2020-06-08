package api

import (
	"net/http"

	internalHTTP "github.com/artii15/termination-detector/pkg/http"
	log "github.com/sirupsen/logrus"
)

type RequestsHandlersMap map[internalHTTP.ResourcePath]map[internalHTTP.Method]RequestHandler

type RequestHandler interface {
	HandleRequest(request internalHTTP.Request) (internalHTTP.Response, error)
}

type Router struct {
	requestsHandlers RequestsHandlersMap
}

func NewRouter(requestsHandlers RequestsHandlersMap) *Router {
	return &Router{
		requestsHandlers: requestsHandlers,
	}
}

func (router *Router) Route(request internalHTTP.Request) internalHTTP.Response {
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

func CreateDefaultTextResponseWithStatus(statusCode int) internalHTTP.Response {
	return internalHTTP.Response{
		StatusCode: statusCode,
		Body:       http.StatusText(statusCode),
		Headers: map[string]string{
			internalHTTP.ContentTypeHeaderName: internalHTTP.ContentTypeTextPlain,
		},
	}
}
