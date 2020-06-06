package client

import (
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	internalHTTP "github.com/nordcloud/termination-detector/pkg/http"
	"github.com/pkg/errors"
)

type RequestModifier interface {
	ModifyRequest(request *http.Request) error
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
	var payloadReader io.Reader
	if len(request.Body) > 0 {
		payloadReader = strings.NewReader(request.Body)
	}
	httpRequest, err := http.NewRequest(string(request.Method), request.FullURL(executor.baseURL), payloadReader)
	if err != nil {
		return internalHTTP.Response{}, errors.Wrapf(err, "failed to build request: %+v", request)
	}

	for _, requestModifier := range executor.requestModifiers {
		if err := requestModifier.ModifyRequest(httpRequest); err != nil {
			return internalHTTP.Response{}, errors.Wrapf(err, "failed to build request: %+v", request)
		}
	}

	response, err := executor.requestDoer.Do(httpRequest)
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

func readHeaders(header http.Header) (headers map[string]string) {
	headers = make(map[string]string)
	for headerName := range header {
		headers[headerName] = header.Get(headerName)
	}
	return headers
}

func readResponseBody(body io.ReadCloser) (string, error) {
	responseBytes, err := ioutil.ReadAll(body)
	if err != nil {
		return "", err
	}
	return string(responseBytes), nil
}
