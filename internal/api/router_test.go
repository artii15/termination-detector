package api_test

import (
	"net/http"
	"testing"

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

func (handler *requestHandlerMock) HandleRequest(request api.Request) (api.Response, error) {
	args := handler.Called(request)
	return args.Get(0).(api.Response), args.Error(1)
}

func newRouterWithMocks() routerWithMocks {
	getTaskRequestHandler := new(requestHandlerMock)
	requestsHandlers := api.RequestsHandlersMap{
		api.ResourcePathTask: {
			api.HTTPMethodGet: getTaskRequestHandler,
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

	request := api.Request{
		ResourcePath: api.ResourcePathTask,
		HTTPMethod:   api.HTTPMethodGet,
	}
	expectedResponse := api.Response{
		StatusCode: http.StatusOK,
		Body:       http.StatusText(http.StatusOK),
	}
	routerAndMocks.getTaskHandler.On("HandleRequest", request).Return(expectedResponse, nil)

	response := routerAndMocks.router.Route(request)
	assert.Equal(t, expectedResponse, response)
}

func TestRouter_Route_UnknownResource(t *testing.T) {
	routerAndMocks := newRouterWithMocks()

	request := api.Request{
		ResourcePath: "unknown",
		HTTPMethod:   http.MethodGet,
	}
	expectedResponse := api.Response{
		StatusCode: http.StatusNotFound,
		Body:       http.StatusText(http.StatusNotFound),
		Headers: map[string]string{
			api.ContentTypeHeaderName: api.ContentTypeTextPlain,
		},
	}

	response := routerAndMocks.router.Route(request)
	assert.Equal(t, expectedResponse, response)
}

func TestRouter_Route_UnknownMethod(t *testing.T) {
	routerAndMocks := newRouterWithMocks()

	request := api.Request{
		ResourcePath: api.ResourcePathTask,
		HTTPMethod:   "PATCH",
	}
	expectedResponse := api.Response{
		StatusCode: http.StatusMethodNotAllowed,
		Body:       http.StatusText(http.StatusMethodNotAllowed),
		Headers: map[string]string{
			api.ContentTypeHeaderName: api.ContentTypeTextPlain,
		},
	}

	response := routerAndMocks.router.Route(request)
	assert.Equal(t, expectedResponse, response)
}

func TestRouter_Route_HandlerError(t *testing.T) {
	routerAndMocks := newRouterWithMocks()

	request := api.Request{
		ResourcePath: api.ResourcePathTask,
		HTTPMethod:   api.HTTPMethodGet,
	}
	expectedError := errors.New("error")
	routerAndMocks.getTaskHandler.On("HandleRequest", request).Return(api.Response{}, expectedError)

	expectedResponse := api.Response{
		StatusCode: http.StatusInternalServerError,
		Body:       http.StatusText(http.StatusInternalServerError),
		Headers: map[string]string{
			api.ContentTypeHeaderName: api.ContentTypeTextPlain,
		},
	}
	response := routerAndMocks.router.Route(request)
	assert.Equal(t, expectedResponse, response)
}
