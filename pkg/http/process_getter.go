package http

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/artii15/termination-detector/pkg/process"
)

type ProcessGetter struct {
	requestExecutor requestExecutor
}

func NewProcessGetter(requestExecutor requestExecutor) *ProcessGetter {
	return &ProcessGetter{
		requestExecutor: requestExecutor,
	}
}

func (getter *ProcessGetter) Get(processID string) (*process.Process, error) {
	response, err := getter.requestExecutor.ExecuteRequest(Request{
		Method:       MethodGet,
		ResourcePath: ResourcePathProcess,
		PathParameters: map[PathParameter]string{
			PathParameterProcessID: processID,
		},
	})
	if err != nil || response.StatusCode == http.StatusNotFound {
		return nil, err
	}
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected error occurred: %d %s", response.StatusCode, response.Body)
	}

	var proc Process
	err = json.Unmarshal([]byte(response.Body), &proc)
	return proc.optionalInternalProcess(), err
}
