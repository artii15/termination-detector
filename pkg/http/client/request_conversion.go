package client

import (
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	internalHTTP "github.com/artii15/termination-detector/pkg/http"
)

func InternalRequestToNative(baseURL string, request internalHTTP.Request) (*http.Request, error) {
	var payloadReader io.Reader
	if len(request.Body) > 0 {
		payloadReader = strings.NewReader(request.Body)
	}
	return http.NewRequest(string(request.Method), request.FullURL(baseURL), payloadReader)
}

func readHeaders(header http.Header) (headers map[string]string) {
	headers = make(map[string]string)
	for headerName := range header {
		headers[headerName] = header.Get(headerName)
	}
	return headers
}

func readResponseBody(body io.ReadCloser) (string, error) {
	if body == nil {
		return "", nil
	}
	responseBytes, err := ioutil.ReadAll(body)
	if err != nil {
		return "", err
	}
	return string(responseBytes), nil
}
