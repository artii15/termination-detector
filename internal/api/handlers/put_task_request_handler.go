package handlers

import (
	"fmt"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/nordcloud/termination-detector/internal/api"
	"github.com/nordcloud/termination-detector/internal/task"
)

const TaskInTerminalStateErrorMessage = "task is in terminal state and can not be updated"

type PutTaskRequestHandler struct {
	registerer task.Registerer
}

func NewPutTaskRequestHandler(registerer task.Registerer) *PutTaskRequestHandler {
	return &PutTaskRequestHandler{
		registerer: registerer,
	}
}

func (handler *PutTaskRequestHandler) HandleRequest(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	unmarshalledTask, err := api.UnmarshalTask(request.Body)
	if err != nil {
		return events.APIGatewayProxyResponse{}, err
	}

	registrationResult, err := handler.registerer.Register(task.RegistrationData{
		ID:             request.PathParameters[api.TaskIDPathParameter],
		ProcessID:      request.PathParameters[api.ProcessIDPathParameter],
		ExpirationTime: unmarshalledTask.ExpirationTime,
	})
	if err != nil {
		return events.APIGatewayProxyResponse{}, err
	}

	return mapTaskRegistrationStatusToResponse(request, registrationResult)
}

func mapTaskRegistrationStatusToResponse(request events.APIGatewayProxyRequest,
	registrationResult task.RegistrationResult) (events.APIGatewayProxyResponse, error) {
	switch registrationResult {
	case task.RegistrationResultCreated:
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusCreated,
			Headers:    map[string]string{api.ContentTypeHeaderName: api.ContentTypeApplicationJSON},
			Body:       request.Body,
		}, nil
	case task.RegistrationResultChanged:
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusOK,
			Headers:    map[string]string{api.ContentTypeHeaderName: api.ContentTypeApplicationJSON},
			Body:       request.Body,
		}, nil
	case task.RegistrationResultNotChanged:
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusNoContent,
		}, nil
	case task.RegistrationResultErrorTerminalState:
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusConflict,
			Headers:    map[string]string{api.ContentTypeHeaderName: api.ContentTypeTextPlain},
			Body:       TaskInTerminalStateErrorMessage,
		}, nil
	default:
		return events.APIGatewayProxyResponse{}, fmt.Errorf("unknown registration result: %s", registrationResult)
	}
}
