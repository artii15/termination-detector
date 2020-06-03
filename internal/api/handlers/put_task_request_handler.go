package handlers

import (
	"fmt"
	"net/http"

	"github.com/nordcloud/termination-detector/internal/api"
	internalHTTP "github.com/nordcloud/termination-detector/pkg/http"
	"github.com/nordcloud/termination-detector/pkg/task"
)

const (
	TaskAlreadyCreatedErrorMessage = "task already created"
	InvalidPayloadErrorMessage     = "invalid payload provided"
)

type PutTaskRequestHandler struct {
	registerer task.Registerer
}

func NewPutTaskRequestHandler(registerer task.Registerer) *PutTaskRequestHandler {
	return &PutTaskRequestHandler{
		registerer: registerer,
	}
}

func (handler *PutTaskRequestHandler) HandleRequest(request internalHTTP.Request) (internalHTTP.Response, error) {
	unmarshalledTask, err := internalHTTP.UnmarshalTask(request.Body)
	if err != nil {
		return internalHTTP.Response{
			StatusCode: http.StatusBadRequest,
			Body:       InvalidPayloadErrorMessage,
			Headers:    map[string]string{api.ContentTypeHeaderName: api.ContentTypeTextPlain},
		}, nil
	}

	registrationResult, err := handler.registerer.Register(task.RegistrationData{
		ID: task.ID{
			ProcessID: request.PathParameters[internalHTTP.PathParameterProcessID],
			TaskID:    request.PathParameters[internalHTTP.PathParameterTaskID],
		},
		ExpirationTime: unmarshalledTask.ExpirationTime,
	})
	if err != nil {
		return internalHTTP.Response{}, err
	}

	return mapTaskRegistrationStatusToResponse(request, registrationResult)
}

func mapTaskRegistrationStatusToResponse(request internalHTTP.Request,
	registrationResult task.RegistrationResult) (internalHTTP.Response, error) {
	switch registrationResult {
	case task.RegistrationResultCreated:
		return internalHTTP.Response{
			StatusCode: http.StatusCreated,
			Headers:    map[string]string{api.ContentTypeHeaderName: api.ContentTypeApplicationJSON},
			Body:       request.Body,
		}, nil
	case task.RegistrationResultAlreadyRegistered:
		return internalHTTP.Response{
			StatusCode: http.StatusConflict,
			Headers:    map[string]string{api.ContentTypeHeaderName: api.ContentTypeTextPlain},
			Body:       TaskAlreadyCreatedErrorMessage,
		}, nil
	default:
		return internalHTTP.Response{}, fmt.Errorf("unknown registration result: %s", registrationResult)
	}
}
