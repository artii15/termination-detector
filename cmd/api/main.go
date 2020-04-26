package main

import (
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/nordcloud/termination-detector/internal/api"
)

func main() {
	router := api.NewRouter(map[api.ResourcePath]map[api.HTTPMethod]api.RequestHandler{})
	lambda.Start(router.Route)
}
