package handlers

import (
	"net/http"

	"github.com/nordcloud/termination-detector/internal/api"
	internalHTTP "github.com/nordcloud/termination-detector/pkg/http"
	"github.com/nordcloud/termination-detector/pkg/process"
)

type GetProcessRequestHandler struct {
	processGetter process.Getter
}

func NewGetProcessRequestHandler(processGetter process.Getter) *GetProcessRequestHandler {
	return &GetProcessRequestHandler{
		processGetter: processGetter,
	}
}

func (handler *GetProcessRequestHandler) HandleRequest(request internalHTTP.Request) (internalHTTP.Response, error) {
	foundProcess, err := handler.processGetter.Get(request.PathParameters[internalHTTP.PathParameterProcessID])
	if err != nil {
		return internalHTTP.Response{}, err
	}
	if foundProcess == nil {
		return api.CreateDefaultTextResponseWithStatus(http.StatusNotFound), nil
	}

	return internalHTTP.Response{
		StatusCode: http.StatusOK,
		Body:       api.ConvertInternalProcess(*foundProcess).JSON(),
		Headers:    map[string]string{api.ContentTypeHeaderName: api.ContentTypeApplicationJSON},
	}, nil
}
