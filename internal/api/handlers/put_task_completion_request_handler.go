package handlers

import (
	"github.com/aws/aws-lambda-go/events"
	"github.com/nordcloud/termination-detector/internal/task"
)

type PutTaskCompletionRequestHandler struct {
	completer task.Completer
}

func NewPutTaskCompletionRequestHandler(completer task.Completer) *PutTaskCompletionRequestHandler {
	return &PutTaskCompletionRequestHandler{
		completer: completer,
	}
}

func (handler *PutTaskCompletionRequestHandler) HandleRequest(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	return events.APIGatewayProxyResponse{}, nil
}
