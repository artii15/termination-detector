package api

import "github.com/aws/aws-lambda-go/events"

type GetTaskRequestHandler struct {
}

func CreateGetTaskRequestHandler() *GetTaskRequestHandler {
	return &GetTaskRequestHandler{}
}

func (handler *GetTaskRequestHandler) HandleRequest(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	return events.APIGatewayProxyResponse{}, nil
}
