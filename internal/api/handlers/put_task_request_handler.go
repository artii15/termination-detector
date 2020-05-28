package handlers

import (
	"fmt"
	"net/http"

	"github.com/nordcloud/termination-detector/internal/api"
	"github.com/nordcloud/termination-detector/internal/task"
)

const (
	TaskInIncompatibleStateErrorMessage = "task is in incompatible state and can not be updated"
	InvalidPayloadErrorMessage          = "invalid payload provided"
)

type PutTaskRequestHandler struct {
	registerer task.Registerer
}

func NewPutTaskRequestHandler(registerer task.Registerer) *PutTaskRequestHandler {
	return &PutTaskRequestHandler{
		registerer: registerer,
	}
}

func (handler *PutTaskRequestHandler) HandleRequest(request api.Request) (api.Response, error) {
	unmarshalledTask, err := api.UnmarshalTask(request.Body)
	if err != nil {
		return api.Response{
			StatusCode: http.StatusBadRequest,
			Body:       InvalidPayloadErrorMessage,
			Headers:    map[string]string{api.ContentTypeHeaderName: api.ContentTypeTextPlain},
		}, nil
	}

	registrationResult, err := handler.registerer.Register(task.RegistrationData{
		ID: task.ID{
			ProcessID: request.PathParameters[api.ProcessIDPathParameter],
			TaskID:    request.PathParameters[api.TaskIDPathParameter],
		},
		ExpirationTime: unmarshalledTask.ExpirationTime,
	})
	if err != nil {
		return api.Response{}, err
	}

	return mapTaskRegistrationStatusToResponse(request, registrationResult)
}

func mapTaskRegistrationStatusToResponse(request api.Request,
	registrationResult task.RegistrationResult) (api.Response, error) {
	switch registrationResult {
	case task.RegistrationResultCreated:
		return api.Response{
			StatusCode: http.StatusCreated,
			Headers:    map[string]string{api.ContentTypeHeaderName: api.ContentTypeApplicationJSON},
			Body:       request.Body,
		}, nil
	case task.RegistrationResultAlreadyRegistered:
		return api.Response{
			StatusCode: http.StatusConflict,
			Headers:    map[string]string{api.ContentTypeHeaderName: api.ContentTypeTextPlain},
			Body:       TaskInIncompatibleStateErrorMessage,
		}, nil
	default:
		return api.Response{}, fmt.Errorf("unknown registration result: %s", registrationResult)
	}
}
