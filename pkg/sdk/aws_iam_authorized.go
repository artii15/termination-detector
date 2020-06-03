package sdk

import (
	"github.com/aws/aws-sdk-go/aws/credentials"
	v4 "github.com/aws/aws-sdk-go/aws/signer/v4"
	"github.com/nordcloud/termination-detector/pkg/http"
	"github.com/nordcloud/termination-detector/pkg/http/client"
	"github.com/nordcloud/termination-detector/pkg/process"
	"github.com/nordcloud/termination-detector/pkg/task"
)

type SDK struct {
	processGetter  process.Getter
	taskRegisterer task.Registerer
	taskCompleter  task.Completer
}

func NewAWSIAMAuthorized(apiURL, region string, awsCredentials *credentials.Credentials) *SDK {
	requestSigner := v4.NewSigner(awsCredentials)
	iamAuthorizingModifier := client.NewIAMAuthorizingModifier(requestSigner, region)

	return New(apiURL, iamAuthorizingModifier)
}

func New(apiURL string, requestModifiers ...client.RequestModifier) *SDK {
	requestExecutor := client.New(apiURL, requestModifiers...)
	return &SDK{
		processGetter:  http.NewProcessGetter(requestExecutor),
		taskRegisterer: http.NewTaskRegisterer(requestExecutor),
		taskCompleter:  http.NewTaskCompleter(requestExecutor),
	}
}
