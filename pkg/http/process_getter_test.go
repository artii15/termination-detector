package http_test

import (
	"errors"
	"net/http"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	internalHTTP "github.com/nordcloud/termination-detector/pkg/http"
	"github.com/nordcloud/termination-detector/pkg/process"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type requestExecutorMock struct {
	mock.Mock
}

func (executor *requestExecutorMock) ExecuteRequest(request internalHTTP.Request) (internalHTTP.Response, error) {
	args := executor.Called(request)
	return args.Get(0).(internalHTTP.Response), args.Error(1)
}

type processGetterWithMocks struct {
	requestExecutor *requestExecutorMock
	procGetter      *internalHTTP.ProcessGetter
}

func newProcessGetterWithMocks() *processGetterWithMocks {
	requestExecutor := new(requestExecutorMock)
	return &processGetterWithMocks{
		requestExecutor: requestExecutor,
		procGetter:      internalHTTP.NewProcessGetter(requestExecutor),
	}
}

func TestProcessGetter_Get(t *testing.T) {
	procGetterAndMocks := newProcessGetterWithMocks()
	processToGet := process.Process{
		ID:           "1",
		State:        process.StateError,
		StateMessage: aws.String("failed"),
	}
	httpProcessToGet := internalHTTP.ConvertInternalToHTTPProcess(processToGet)

	procGetterAndMocks.requestExecutor.On("ExecuteRequest", internalHTTP.Request{
		Method:       internalHTTP.MethodGet,
		ResourcePath: internalHTTP.ResourcePathProcess,
		PathParameters: map[internalHTTP.PathParameter]string{
			internalHTTP.PathParameterProcessID: processToGet.ID,
		},
	}).Return(internalHTTP.Response{
		StatusCode: http.StatusOK,
		Body:       httpProcessToGet.JSON(),
		Headers: map[string]string{
			internalHTTP.ContentTypeHeaderName: internalHTTP.ContentTypeApplicationJSON,
		},
	}, nil)

	proc, err := procGetterAndMocks.procGetter.Get(processToGet.ID)
	assert.NoError(t, err)
	assert.NotNil(t, proc)
	assert.Equal(t, processToGet, *proc)
}

func TestProcessGetter_Get_ProcessNotFound(t *testing.T) {
	procGetterAndMocks := newProcessGetterWithMocks()
	procID := "1"

	procGetterAndMocks.requestExecutor.On("ExecuteRequest", internalHTTP.Request{
		Method:       internalHTTP.MethodGet,
		ResourcePath: internalHTTP.ResourcePathProcess,
		PathParameters: map[internalHTTP.PathParameter]string{
			internalHTTP.PathParameterProcessID: procID,
		},
	}).Return(internalHTTP.Response{
		StatusCode: http.StatusNotFound,
		Headers: map[string]string{
			internalHTTP.ContentTypeHeaderName: internalHTTP.ContentTypeApplicationJSON,
		},
	}, nil)

	proc, err := procGetterAndMocks.procGetter.Get(procID)
	assert.NoError(t, err)
	assert.Nil(t, proc)
}

func TestProcessGetter_Get_RequestExecutorError(t *testing.T) {
	procGetterAndMocks := newProcessGetterWithMocks()
	procID := "1"

	procGetterAndMocks.requestExecutor.On("ExecuteRequest", internalHTTP.Request{
		Method:       internalHTTP.MethodGet,
		ResourcePath: internalHTTP.ResourcePathProcess,
		PathParameters: map[internalHTTP.PathParameter]string{
			internalHTTP.PathParameterProcessID: procID,
		},
	}).Return(internalHTTP.Response{
		StatusCode: http.StatusOK,
		Headers: map[string]string{
			internalHTTP.ContentTypeHeaderName: internalHTTP.ContentTypeApplicationJSON,
		},
	}, errors.New("error"))

	_, err := procGetterAndMocks.procGetter.Get(procID)
	assert.Error(t, err)
}

func TestProcessGetter_Get_UnknownResponseStatus(t *testing.T) {
	procGetterAndMocks := newProcessGetterWithMocks()
	procID := "1"

	procGetterAndMocks.requestExecutor.On("ExecuteRequest", internalHTTP.Request{
		Method:       internalHTTP.MethodGet,
		ResourcePath: internalHTTP.ResourcePathProcess,
		PathParameters: map[internalHTTP.PathParameter]string{
			internalHTTP.PathParameterProcessID: procID,
		},
	}).Return(internalHTTP.Response{
		StatusCode: http.StatusInternalServerError,
		Headers: map[string]string{
			internalHTTP.ContentTypeHeaderName: internalHTTP.ContentTypeApplicationJSON,
		},
	}, nil)

	_, err := procGetterAndMocks.procGetter.Get(procID)
	assert.Error(t, err)
}
