package api_test

import (
	"net/http"
	"testing"

	"github.com/aws/aws-lambda-go/events"
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

func (handler *requestHandlerMock) HandleRequest(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	args := handler.Called(request)
	return args.Get(0).(events.APIGatewayProxyResponse), args.Error(1)
}

func TestRouter_Route(t *testing.T) {
	routerAndMocks := newRouterWithMocks()

	request := events.APIGatewayProxyRequest{
		Resource:   string(api.ResourcePathTask),
		HTTPMethod: string(api.HTTPMethodGet),
	}
	expectedResponse := events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Body:       http.StatusText(http.StatusOK),
	}
	routerAndMocks.getTaskHandler.On("HandleRequest", request).Return(expectedResponse, nil)

	response := routerAndMocks.router.Route(request)
	assert.Equal(t, expectedResponse, response)
}

func TestRouter_Route_UnknownResource(t *testing.T) {
	routerAndMocks := newRouterWithMocks()

	request := events.APIGatewayProxyRequest{
		Resource:   "unknown",
		HTTPMethod: http.MethodGet,
	}
	expectedResponse := events.APIGatewayProxyResponse{
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

	request := events.APIGatewayProxyRequest{
		Resource:   string(api.ResourcePathTask),
		HTTPMethod: "PATCH",
	}
	expectedResponse := events.APIGatewayProxyResponse{
		StatusCode: http.StatusNotFound,
		Body:       http.StatusText(http.StatusNotFound),
		Headers: map[string]string{
			api.ContentTypeHeaderName: api.ContentTypeTextPlain,
		},
	}

	response := routerAndMocks.router.Route(request)
	assert.Equal(t, expectedResponse, response)
}

func TestRouter_Route_HandlerError(t *testing.T) {
	routerAndMocks := newRouterWithMocks()

	request := events.APIGatewayProxyRequest{
		Resource:   string(api.ResourcePathTask),
		HTTPMethod: string(api.HTTPMethodGet),
	}
	expectedError := errors.New("error")
	routerAndMocks.getTaskHandler.On("HandleRequest", request).Return(events.APIGatewayProxyResponse{}, expectedError)

	expectedResponse := events.APIGatewayProxyResponse{
		StatusCode: http.StatusInternalServerError,
		Body:       http.StatusText(http.StatusInternalServerError),
		Headers: map[string]string{
			api.ContentTypeHeaderName: api.ContentTypeTextPlain,
		},
	}
	response := routerAndMocks.router.Route(request)
	assert.Equal(t, expectedResponse, response)
}
