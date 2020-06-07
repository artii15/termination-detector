package api_test

import (
	"net/http"
	"testing"

	http2 "github.com/nordcloud/termination-detector/pkg/http"

	"github.com/nordcloud/termination-detector/internal/api"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type requestHandlerMock struct {
	mock.Mock
}

type routerWithMocks struct {
	router         *api.Router
	getTaskHandler *requestHandlerMock
}

func (handler *requestHandlerMock) HandleRequest(request http2.Request) (http2.Response, error) {
	args := handler.Called(request)
	return args.Get(0).(http2.Response), args.Error(1)
}

func newRouterWithMocks() routerWithMocks {
	getTaskRequestHandler := new(requestHandlerMock)
	requestsHandlers := api.RequestsHandlersMap{
		http2.ResourcePathTask: {
			http2.MethodGet: getTaskRequestHandler,
		},
	}
	router := api.NewRouter(requestsHandlers)

	return routerWithMocks{
		router:         router,
		getTaskHandler: getTaskRequestHandler,
	}
}

func TestRouter_Route(t *testing.T) {
	routerAndMocks := newRouterWithMocks()

	request := http2.Request{
		ResourcePath: http2.ResourcePathTask,
		Method:       http2.MethodGet,
	}
	expectedResponse := http2.Response{
		StatusCode: http.StatusOK,
		Body:       http.StatusText(http.StatusOK),
	}
	routerAndMocks.getTaskHandler.On("HandleRequest", request).Return(expectedResponse, nil)

	response := routerAndMocks.router.Route(request)
	assert.Equal(t, expectedResponse, response)
}

func TestRouter_Route_UnknownResource(t *testing.T) {
	routerAndMocks := newRouterWithMocks()

	request := http2.Request{
		ResourcePath: "unknown",
		Method:       http.MethodGet,
	}
	expectedResponse := http2.Response{
		StatusCode: http.StatusNotFound,
		Body:       http.StatusText(http.StatusNotFound),
		Headers: map[string]string{
			http2.ContentTypeHeaderName: http2.ContentTypeTextPlain,
		},
	}

	response := routerAndMocks.router.Route(request)
	assert.Equal(t, expectedResponse, response)
}

func TestRouter_Route_UnknownMethod(t *testing.T) {
	routerAndMocks := newRouterWithMocks()

	request := http2.Request{
		ResourcePath: http2.ResourcePathTask,
		Method:       "PATCH",
	}
	expectedResponse := http2.Response{
		StatusCode: http.StatusMethodNotAllowed,
		Body:       http.StatusText(http.StatusMethodNotAllowed),
		Headers: map[string]string{
			http2.ContentTypeHeaderName: http2.ContentTypeTextPlain,
		},
	}

	response := routerAndMocks.router.Route(request)
	assert.Equal(t, expectedResponse, response)
}

func TestRouter_Route_HandlerError(t *testing.T) {
	routerAndMocks := newRouterWithMocks()

	request := http2.Request{
		ResourcePath: http2.ResourcePathTask,
		Method:       http2.MethodGet,
	}
	expectedError := errors.New("error")
	routerAndMocks.getTaskHandler.On("HandleRequest", request).Return(http2.Response{}, expectedError)

	expectedResponse := http2.Response{
		StatusCode: http.StatusInternalServerError,
		Body:       http.StatusText(http.StatusInternalServerError),
		Headers: map[string]string{
			http2.ContentTypeHeaderName: http2.ContentTypeTextPlain,
		},
	}
	response := routerAndMocks.router.Route(request)
	assert.Equal(t, expectedResponse, response)
}
