package main

import (
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/nordcloud/termination-detector/internal/api"
)

func main() {
	router := api.CreateRouter(map[api.ResourcePath]map[api.HTTPMethod]api.RequestHandler{
		api.ResourcePathTask: {
			api.HTTPMethodGet: api.CreateGetTaskRequestHandler(),
		},
	})
	lambda.Start(router.Route)
}
