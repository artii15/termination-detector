package client

import (
	"net/http"

	internalHTTP "github.com/artii15/termination-detector/pkg/http"
	"github.com/pkg/errors"
)

type RequestModifier interface {
	ModifyRequest(request *http.Request) error
}

type RequestToNativeConverter interface {
	ConvertRequestToNative()
}

type HTTPRequestDoer interface {
	Do(request *http.Request) (*http.Response, error)
}

type Client struct {
	requestDoer      HTTPRequestDoer
	baseURL          string
	requestModifiers []RequestModifier
}

func New(httpRequestDoer HTTPRequestDoer, baseURL string, requestModifiers ...RequestModifier) *Client {
	return &Client{
		requestDoer:      httpRequestDoer,
		baseURL:          baseURL,
		requestModifiers: requestModifiers,
	}
}

func (executor *Client) ExecuteRequest(request internalHTTP.Request) (internalHTTP.Response, error) {
	httpRequest, err := executor.buildHTTPRequest(request)
	if err != nil {
		return internalHTTP.Response{}, err
	}

	return executor.executeNativeRequest(httpRequest)
}

func (executor *Client) buildHTTPRequest(request internalHTTP.Request) (*http.Request, error) {
	httpRequest, err := InternalRequestToNative(executor.baseURL, request)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to convert request: %+v", request)
	}

	for _, requestModifier := range executor.requestModifiers {
		if err := requestModifier.ModifyRequest(httpRequest); err != nil {
			return nil, errors.Wrapf(err, "failed to modify request: %+v", httpRequest)
		}
	}
	return httpRequest, nil
}

func (executor *Client) executeNativeRequest(request *http.Request) (internalHTTP.Response, error) {
	response, err := executor.requestDoer.Do(request)
	if err != nil {
		return internalHTTP.Response{}, errors.Wrapf(err, "failed to execute request: %+v", request)
	}
	responseBody, err := readResponseBody(response.Body)
	if err != nil {
		return internalHTTP.Response{}, errors.Wrapf(err, "failed to read response body: %+v", response)
	}
	return internalHTTP.Response{
		StatusCode: response.StatusCode,
		Body:       responseBody,
		Headers:    readHeaders(response.Header),
	}, nil
}
