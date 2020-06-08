package handlers

import (
	"net/http"

	internalHTTP "github.com/artii15/termination-detector/pkg/http"
	"github.com/artii15/termination-detector/pkg/process"
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
		return internalHTTP.CreateDefaultTextResponseWithStatus(http.StatusNotFound), nil
	}

	return internalHTTP.Response{
		StatusCode: http.StatusOK,
		Body:       internalHTTP.ConvertInternalToHTTPProcess(*foundProcess).JSON(),
		Headers:    map[string]string{internalHTTP.ContentTypeHeaderName: internalHTTP.ContentTypeApplicationJSON},
	}, nil
}
