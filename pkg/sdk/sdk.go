package sdk

import (
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go/aws/credentials"
	v4 "github.com/aws/aws-sdk-go/aws/signer/v4"
	internalHTTP "github.com/nordcloud/termination-detector/pkg/http"
	"github.com/nordcloud/termination-detector/pkg/http/client"
	"github.com/nordcloud/termination-detector/pkg/process"
	"github.com/nordcloud/termination-detector/pkg/task"
)

type SDK struct {
	processGetter  process.Getter
	taskRegisterer task.Registerer
	taskCompleter  task.Completer
}

func (sdk *SDK) Get(processID string) (*process.Process, error) {
	return sdk.processGetter.Get(processID)
}

func (sdk *SDK) Register(registrationData task.RegistrationData) (task.RegistrationResult, error) {
	return sdk.taskRegisterer.Register(registrationData)
}

func (sdk *SDK) Complete(request task.CompleteRequest) (task.CompletingResult, error) {
	return sdk.taskCompleter.Complete(request)
}

func NewAWSIAMAuthorized(requestsTimeout time.Duration, apiURL, region string, awsCredentials *credentials.Credentials) *SDK {
	requestSigner := v4.NewSigner(awsCredentials)
	iamAuthorizingModifier := client.NewIAMAuthorizingModifier(requestSigner, region)

	return New(requestsTimeout, apiURL, iamAuthorizingModifier)
}

func New(requestsTimeout time.Duration, apiURL string, requestModifiers ...client.RequestModifier) *SDK {
	httpClient := &http.Client{
		Timeout: requestsTimeout,
	}
	requestExecutor := client.New(httpClient, apiURL, requestModifiers...)
	return &SDK{
		processGetter:  internalHTTP.NewProcessGetter(requestExecutor),
		taskRegisterer: internalHTTP.NewTaskRegisterer(requestExecutor),
		taskCompleter:  internalHTTP.NewTaskCompleter(requestExecutor),
	}
}
