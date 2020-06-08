package http_test

import (
	"net/http"
	"testing"

	internalHTTP "github.com/artii15/termination-detector/pkg/http"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type requestHandlerMock struct {
	mock.Mock
}

type routerWithMocks struct {
	router         *internalHTTP.Router
	getTaskHandler *requestHandlerMock
}

func (handler *requestHandlerMock) HandleRequest(request internalHTTP.Request) (internalHTTP.Response, error) {
	args := handler.Called(request)
	return args.Get(0).(internalHTTP.Response), args.Error(1)
}

func newRouterWithMocks() routerWithMocks {
	getTaskRequestHandler := new(requestHandlerMock)
	requestsHandlers := internalHTTP.RequestsHandlersMap{
		internalHTTP.ResourcePathTask: {
			internalHTTP.MethodGet: getTaskRequestHandler,
		},
	}
	router := internalHTTP.NewRouter(requestsHandlers)

	return routerWithMocks{
		router:         router,
		getTaskHandler: getTaskRequestHandler,
	}
}

func TestRouter_Route(t *testing.T) {
	routerAndMocks := newRouterWithMocks()

	request := internalHTTP.Request{
		ResourcePath: internalHTTP.ResourcePathTask,
		Method:       internalHTTP.MethodGet,
	}
	expectedResponse := internalHTTP.Response{
		StatusCode: http.StatusOK,
		Body:       http.StatusText(http.StatusOK),
	}
	routerAndMocks.getTaskHandler.On("HandleRequest", request).Return(expectedResponse, nil)

	response := routerAndMocks.router.Route(request)
	assert.Equal(t, expectedResponse, response)
}

func TestRouter_Route_UnknownResource(t *testing.T) {
	routerAndMocks := newRouterWithMocks()

	request := internalHTTP.Request{
		ResourcePath: "unknown",
		Method:       http.MethodGet,
	}
	expectedResponse := internalHTTP.Response{
		StatusCode: http.StatusNotFound,
		Body:       http.StatusText(http.StatusNotFound),
		Headers: map[string]string{
			internalHTTP.ContentTypeHeaderName: internalHTTP.ContentTypeTextPlain,
		},
	}

	response := routerAndMocks.router.Route(request)
	assert.Equal(t, expectedResponse, response)
}

func TestRouter_Route_UnknownMethod(t *testing.T) {
	routerAndMocks := newRouterWithMocks()

	request := internalHTTP.Request{
		ResourcePath: internalHTTP.ResourcePathTask,
		Method:       "PATCH",
	}
	expectedResponse := internalHTTP.Response{
		StatusCode: http.StatusMethodNotAllowed,
		Body:       http.StatusText(http.StatusMethodNotAllowed),
		Headers: map[string]string{
			internalHTTP.ContentTypeHeaderName: internalHTTP.ContentTypeTextPlain,
		},
	}

	response := routerAndMocks.router.Route(request)
	assert.Equal(t, expectedResponse, response)
}

func TestRouter_Route_HandlerError(t *testing.T) {
	routerAndMocks := newRouterWithMocks()

	request := internalHTTP.Request{
		ResourcePath: internalHTTP.ResourcePathTask,
		Method:       internalHTTP.MethodGet,
	}
	expectedError := errors.New("error")
	routerAndMocks.getTaskHandler.On("HandleRequest", request).Return(internalHTTP.Response{}, expectedError)

	expectedResponse := internalHTTP.Response{
		StatusCode: http.StatusInternalServerError,
		Body:       http.StatusText(http.StatusInternalServerError),
		Headers: map[string]string{
			internalHTTP.ContentTypeHeaderName: internalHTTP.ContentTypeTextPlain,
		},
	}
	response := routerAndMocks.router.Route(request)
	assert.Equal(t, expectedResponse, response)
}
