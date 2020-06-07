package client_test

import (
	"bytes"
	"errors"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
	"time"

	internalHTTP "github.com/nordcloud/termination-detector/pkg/http"
	"github.com/nordcloud/termination-detector/pkg/http/client"
	"github.com/nordcloud/termination-detector/pkg/process"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type requestDoerMock struct {
	mock.Mock
}

func (doer *requestDoerMock) Do(request *http.Request) (*http.Response, error) {
	args := doer.Called(request)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*http.Response), args.Error(1)
}

type requestModifierMock struct {
	mock.Mock
}

func (modifier *requestModifierMock) ModifyRequest(request *http.Request) error {
	return modifier.Called(request).Error(0)
}

type clientWithMocks struct {
	client          *client.Client
	requestDoer     *requestDoerMock
	apiURL          string
	requestModifier *requestModifierMock
}

func (clientAndMocks *clientWithMocks) assertExpectations(t *testing.T) {
	clientAndMocks.requestModifier.AssertExpectations(t)
	clientAndMocks.requestDoer.AssertExpectations(t)
}

func newClientWithMocks() *clientWithMocks {
	requestDoer := new(requestDoerMock)
	requestModifier := new(requestModifierMock)
	apiURL := "https://test.com"
	return &clientWithMocks{
		client:          client.New(requestDoer, apiURL, requestModifier),
		requestDoer:     requestDoer,
		apiURL:          apiURL,
		requestModifier: requestModifier,
	}
}

func TestClient_ExecuteRequest(t *testing.T) {
	clientAndMocks := newClientWithMocks()
	processID := "1"
	request := internalHTTP.Request{
		Method:       internalHTTP.MethodGet,
		ResourcePath: internalHTTP.ResourcePathProcess,
		PathParameters: map[internalHTTP.PathParameter]string{
			internalHTTP.PathParameterProcessID: processID,
		},
	}
	nativeRequest, err := client.InternalRequestToNative(clientAndMocks.apiURL, request)
	assert.NoError(t, err)

	clientAndMocks.requestModifier.On("ModifyRequest", nativeRequest).Return(nil).Once()

	returnedProcess := internalHTTP.Process{ID: processID, State: process.StateCreated}
	returnedProcessJSON := returnedProcess.JSON()
	nativeResponse := &http.Response{
		Status:     http.StatusText(http.StatusOK),
		StatusCode: http.StatusOK,
		Header:     map[string][]string{internalHTTP.ContentTypeHeaderName: {internalHTTP.ContentTypeApplicationJSON}},
		Body:       ioutil.NopCloser(strings.NewReader(returnedProcessJSON)),
	}
	clientAndMocks.requestDoer.On("Do", nativeRequest).Return(nativeResponse, nil)

	response, err := clientAndMocks.client.ExecuteRequest(request)
	assert.NoError(t, err)
	assert.Equal(t, internalHTTP.Response{
		StatusCode: nativeResponse.StatusCode,
		Body:       returnedProcessJSON,
		Headers:    map[string]string{internalHTTP.ContentTypeHeaderName: internalHTTP.ContentTypeApplicationJSON},
	}, response)
	clientAndMocks.assertExpectations(t)
}

func TestClient_ExecuteRequest_NoResponseBody(t *testing.T) {
	clientAndMocks := newClientWithMocks()
	processID := "1"
	request := internalHTTP.Request{
		Method:       internalHTTP.MethodGet,
		ResourcePath: internalHTTP.ResourcePathProcess,
		PathParameters: map[internalHTTP.PathParameter]string{
			internalHTTP.PathParameterProcessID: processID,
		},
	}
	nativeRequest, err := client.InternalRequestToNative(clientAndMocks.apiURL, request)
	assert.NoError(t, err)

	clientAndMocks.requestModifier.On("ModifyRequest", nativeRequest).Return(nil).Once()

	nativeResponse := &http.Response{
		Status:     http.StatusText(http.StatusNotFound),
		StatusCode: http.StatusNotFound,
		Header:     map[string][]string{internalHTTP.ContentTypeHeaderName: {internalHTTP.ContentTypeTextPlain}},
		Body:       nil,
	}
	clientAndMocks.requestDoer.On("Do", nativeRequest).Return(nativeResponse, nil)

	response, err := clientAndMocks.client.ExecuteRequest(request)
	assert.NoError(t, err)
	assert.Equal(t, internalHTTP.Response{
		StatusCode: nativeResponse.StatusCode,
		Body:       "",
		Headers:    map[string]string{internalHTTP.ContentTypeHeaderName: internalHTTP.ContentTypeTextPlain},
	}, response)
	clientAndMocks.assertExpectations(t)
}

func TestClient_ExecuteRequest_RequestWithPayload(t *testing.T) {
	clientAndMocks := newClientWithMocks()
	processID := "1"
	taskID := "2"
	currentTime := time.Now()
	taskRegistrationData := internalHTTP.Task{
		ExpirationTime: currentTime.Add(time.Hour),
	}
	request := internalHTTP.Request{
		Method:       internalHTTP.MethodPut,
		ResourcePath: internalHTTP.ResourcePathTask,
		PathParameters: map[internalHTTP.PathParameter]string{
			internalHTTP.PathParameterProcessID: processID,
			internalHTTP.PathParameterTaskID:    taskID,
		},
		Body: taskRegistrationData.JSON(),
	}
	nativeRequest, err := client.InternalRequestToNative(clientAndMocks.apiURL, request)
	assert.NoError(t, err)
	nativeRequestBodyBytes, err := ioutil.ReadAll(nativeRequest.Body)
	assert.NoError(t, err)
	nativeRequest.Body = ioutil.NopCloser(bytes.NewReader(nativeRequestBodyBytes))
	assert.Equal(t, taskRegistrationData.JSON(), string(nativeRequestBodyBytes))

	clientAndMocks.requestModifier.On("ModifyRequest", mock.MatchedBy(func(httpRequest *http.Request) bool {
		return areRequestsEqual(t, nativeRequest, httpRequest)
	})).Return(nil).Once()

	responseBody := ioutil.NopCloser(strings.NewReader(taskRegistrationData.JSON()))
	nativeResponse := &http.Response{
		Status:     http.StatusText(http.StatusCreated),
		StatusCode: http.StatusCreated,
		Header:     map[string][]string{internalHTTP.ContentTypeHeaderName: {internalHTTP.ContentTypeApplicationJSON}},
		Body:       responseBody,
	}
	clientAndMocks.requestDoer.On("Do", mock.MatchedBy(func(httpRequest *http.Request) bool {
		return areRequestsEqual(t, nativeRequest, httpRequest)
	})).Return(nativeResponse, nil)

	response, err := clientAndMocks.client.ExecuteRequest(request)
	assert.NoError(t, err)
	assert.Equal(t, internalHTTP.Response{
		StatusCode: nativeResponse.StatusCode,
		Body:       taskRegistrationData.JSON(),
		Headers:    map[string]string{internalHTTP.ContentTypeHeaderName: internalHTTP.ContentTypeApplicationJSON},
	}, response)
	clientAndMocks.assertExpectations(t)
}

func areRequestsEqual(t *testing.T, expected, actual *http.Request) bool {
	return expected.Method == actual.Method &&
		assert.ObjectsAreEqual(expected.Header, actual.Header) &&
		areRequestsBodiesEqual(t, expected, actual)

}

func areRequestsBodiesEqual(t *testing.T, expected, actual *http.Request) bool {
	expectedBodyContent, err := ioutil.ReadAll(expected.Body)
	assert.NoError(t, err)
	actualBodyContent, err := ioutil.ReadAll(actual.Body)
	assert.NoError(t, err)

	expected.Body = ioutil.NopCloser(bytes.NewReader(expectedBodyContent))
	actual.Body = ioutil.NopCloser(bytes.NewReader(actualBodyContent))

	return string(expectedBodyContent) == string(actualBodyContent)
}

func TestClient_ExecuteRequest_ModifierError(t *testing.T) {
	clientAndMocks := newClientWithMocks()
	processID := "1"
	request := internalHTTP.Request{
		Method:       internalHTTP.MethodGet,
		ResourcePath: internalHTTP.ResourcePathProcess,
		PathParameters: map[internalHTTP.PathParameter]string{
			internalHTTP.PathParameterProcessID: processID,
		},
	}
	nativeRequest, err := client.InternalRequestToNative(clientAndMocks.apiURL, request)
	assert.NoError(t, err)

	clientAndMocks.requestModifier.On("ModifyRequest", nativeRequest).Return(errors.New("error"))

	_, err = clientAndMocks.client.ExecuteRequest(request)
	assert.Error(t, err)
	clientAndMocks.assertExpectations(t)
}

func TestClient_ExecuteRequest_DoerError(t *testing.T) {
	clientAndMocks := newClientWithMocks()
	processID := "1"
	request := internalHTTP.Request{
		Method:       internalHTTP.MethodGet,
		ResourcePath: internalHTTP.ResourcePathProcess,
		PathParameters: map[internalHTTP.PathParameter]string{
			internalHTTP.PathParameterProcessID: processID,
		},
	}
	nativeRequest, err := client.InternalRequestToNative(clientAndMocks.apiURL, request)
	assert.NoError(t, err)

	clientAndMocks.requestModifier.On("ModifyRequest", nativeRequest).Return(nil).Once()

	returnedProcess := internalHTTP.Process{ID: processID, State: process.StateCreated}
	returnedProcessJSON := returnedProcess.JSON()
	nativeResponse := &http.Response{
		Status:     http.StatusText(http.StatusOK),
		StatusCode: http.StatusOK,
		Header:     map[string][]string{internalHTTP.ContentTypeHeaderName: {internalHTTP.ContentTypeApplicationJSON}},
		Body:       ioutil.NopCloser(strings.NewReader(returnedProcessJSON)),
	}
	clientAndMocks.requestDoer.On("Do", nativeRequest).Return(nativeResponse, errors.New("error"))

	_, err = clientAndMocks.client.ExecuteRequest(request)
	assert.Error(t, err)
	clientAndMocks.assertExpectations(t)
}
