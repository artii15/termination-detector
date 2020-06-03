package http

import (
	"fmt"
	"net/http"
	"strings"
)

type ResourcePath string
type Method string
type PathParameter string

const (
	PathParameterProcessID PathParameter = "process_id"
	PathParameterTaskID    PathParameter = "task_id"

	ResourcePathTask           ResourcePath = "/processes/{process_id}/tasks/{task_id}"
	ResourcePathTaskCompletion ResourcePath = "/processes/{process_id}/tasks/{task_id}/completion"
	ResourcePathProcess        ResourcePath = "/processes/{process_id}"

	MethodGet Method = http.MethodGet
	MethodPut Method = http.MethodPut
)

type Request struct {
	Method         Method
	ResourcePath   ResourcePath
	Body           string
	PathParameters map[PathParameter]string
}

func (request Request) FullURL(baseURL string) string {
	resourceURL := request.resourceURL()
	return strings.Join([]string{
		strings.TrimRight(baseURL, "/"),
		strings.TrimLeft(resourceURL, "/"),
	}, "/")
}

func (request Request) resourceURL() string {
	url := string(request.ResourcePath)
	for paramName, paramValue := range request.PathParameters {
		pathFragmentToReplace := fmt.Sprintf("{%s}", paramName)
		url = strings.ReplaceAll(url, pathFragmentToReplace, paramValue)
	}
	return url
}

type Response struct {
	StatusCode int
	Body       string
	Headers    map[string]string
}

type requestExecutor interface {
	ExecuteRequest(request Request) (Response, error)
}
