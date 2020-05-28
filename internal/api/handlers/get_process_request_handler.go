package handlers

import (
	"net/http"

	"github.com/nordcloud/termination-detector/internal/process"

	"github.com/nordcloud/termination-detector/internal/api"
)

type GetProcessRequestHandler struct {
	processGetter process.Getter
}

func NewGetProcessRequestHandler(processGetter process.Getter) *GetProcessRequestHandler {
	return &GetProcessRequestHandler{
		processGetter: processGetter,
	}
}

func (handler *GetProcessRequestHandler) HandleRequest(request api.Request) (api.Response, error) {
	foundProcess, err := handler.processGetter.Get(request.PathParameters[api.ProcessIDPathParameter])
	if err != nil {
		return api.Response{}, err
	}
	if foundProcess == nil {
		return api.Response{
			StatusCode: http.StatusNotFound,
			Headers:    map[string]string{api.ContentTypeHeaderName: api.ContentTypeTextPlain},
			Body:       http.StatusText(http.StatusNotFound),
		}, nil
	}

	return api.Response{
		StatusCode: http.StatusOK,
		Body:       api.ConvertInternalProcess(*foundProcess).JSON(),
		Headers:    map[string]string{api.ContentTypeHeaderName: api.ContentTypeApplicationJSON},
	}, nil
}
