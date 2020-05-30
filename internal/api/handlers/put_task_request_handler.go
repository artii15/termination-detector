package handlers

import (
	"fmt"
	"net/http"

	task2 "github.com/nordcloud/termination-detector/pkg/task"

	"github.com/nordcloud/termination-detector/internal/api"
)

const (
	TaskAlreadyCreatedErrorMessage = "task already created"
	InvalidPayloadErrorMessage     = "invalid payload provided"
)

type PutTaskRequestHandler struct {
	registerer task2.Registerer
}

func NewPutTaskRequestHandler(registerer task2.Registerer) *PutTaskRequestHandler {
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

	registrationResult, err := handler.registerer.Register(task2.RegistrationData{
		ID: task2.ID{
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
	registrationResult task2.RegistrationResult) (api.Response, error) {
	switch registrationResult {
	case task2.RegistrationResultCreated:
		return api.Response{
			StatusCode: http.StatusCreated,
			Headers:    map[string]string{api.ContentTypeHeaderName: api.ContentTypeApplicationJSON},
			Body:       request.Body,
		}, nil
	case task2.RegistrationResultAlreadyRegistered:
		return api.Response{
			StatusCode: http.StatusConflict,
			Headers:    map[string]string{api.ContentTypeHeaderName: api.ContentTypeTextPlain},
			Body:       TaskAlreadyCreatedErrorMessage,
		}, nil
	default:
		return api.Response{}, fmt.Errorf("unknown registration result: %s", registrationResult)
	}
}
