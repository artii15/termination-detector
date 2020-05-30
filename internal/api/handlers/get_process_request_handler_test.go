package handlers_test

import (
	"errors"
	"net/http"
	"testing"

	"github.com/nordcloud/termination-detector/internal/api/handlers"

	"github.com/aws/aws-sdk-go/aws"

	"github.com/stretchr/testify/assert"

	"github.com/nordcloud/termination-detector/internal/api"
	"github.com/nordcloud/termination-detector/pkg/process"
	"github.com/stretchr/testify/mock"
)

type processGetterMock struct {
	mock.Mock
}

func (getter *processGetterMock) Get(processID string) (*process.Process, error) {
	args := getter.Called(processID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*process.Process), args.Error(1)
}

type getProcessRequestHandlerWithMocks struct {
	handler       *handlers.GetProcessRequestHandler
	processGetter *processGetterMock
	request       api.Request
	processID     string
}

func (getterAndMocks *getProcessRequestHandlerWithMocks) assertExpectations(t *testing.T) {
	getterAndMocks.processGetter.AssertExpectations(t)
}

func newGetProcessRequestHandlerWithMocks() *getProcessRequestHandlerWithMocks {
	processGetter := new(processGetterMock)
	handler := handlers.NewGetProcessRequestHandler(processGetter)
	processID := "2"
	return &getProcessRequestHandlerWithMocks{
		handler:       handler,
		processGetter: processGetter,
		processID:     processID,
		request: api.Request{
			PathParameters: map[string]string{api.ProcessIDPathParameter: processID},
		},
	}
}

func TestGetProcessRequestHandler_HandleRequest(t *testing.T) {
	handlerAndMocks := newGetProcessRequestHandlerWithMocks()
	foundProcess := process.Process{
		ID:           handlerAndMocks.processID,
		State:        process.StateError,
		StateMessage: aws.String("error"),
	}
	handlerAndMocks.processGetter.On("Get", handlerAndMocks.processID).Return(&foundProcess, nil)

	response, err := handlerAndMocks.handler.HandleRequest(handlerAndMocks.request)
	assert.NoError(t, err)
	handlerAndMocks.assertExpectations(t)
	assert.Equal(t, api.Response{
		StatusCode: http.StatusOK,
		Body: api.Process{
			ID:           foundProcess.ID,
			State:        foundProcess.State,
			StateMessage: foundProcess.StateMessage,
		}.JSON(),
		Headers: map[string]string{api.ContentTypeHeaderName: api.ContentTypeApplicationJSON},
	}, response)
}

func TestGetProcessRequestHandler_HandleRequest_ProcessNotFound(t *testing.T) {
	handlerAndMocks := newGetProcessRequestHandlerWithMocks()
	handlerAndMocks.processGetter.On("Get", handlerAndMocks.processID).Return((*process.Process)(nil), nil)

	response, err := handlerAndMocks.handler.HandleRequest(handlerAndMocks.request)
	assert.NoError(t, err)
	handlerAndMocks.assertExpectations(t)
	assert.Equal(t, api.Response{
		StatusCode: http.StatusNotFound,
		Body:       http.StatusText(http.StatusNotFound),
		Headers:    map[string]string{api.ContentTypeHeaderName: api.ContentTypeTextPlain},
	}, response)
}

func TestGetProcessRequestHandler_HandleRequest_ProcessGetterError(t *testing.T) {
	handlerAndMocks := newGetProcessRequestHandlerWithMocks()
	handlerAndMocks.processGetter.On("Get", handlerAndMocks.processID).
		Return((*process.Process)(nil), errors.New("error"))

	_, err := handlerAndMocks.handler.HandleRequest(handlerAndMocks.request)
	assert.Error(t, err)
	handlerAndMocks.assertExpectations(t)
}
