package client

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	v4 "github.com/aws/aws-sdk-go/aws/signer/v4"
	"github.com/pkg/errors"
)

const apiGatewaySvcNameForSignature = "execute-api"

type IAMAuthorizingModifier struct {
	signer          *v4.Signer
	signatureRegion string
}

func NewIAMAuthorizingModifier(signer *v4.Signer, signatureRegion string) *IAMAuthorizingModifier {
	return &IAMAuthorizingModifier{
		signer:          signer,
		signatureRegion: signatureRegion,
	}
}

func (modifier *IAMAuthorizingModifier) ModifyRequest(request *http.Request) error {
	if err := modifier.signRequest(request); err != nil {
		return errors.Wrap(err, "failed to generate v4 signature for request")
	}
	return nil
}

func (modifier *IAMAuthorizingModifier) signRequest(request *http.Request) error {
	requestContent, err := readRequestBody(request.Body)
	if err != nil {
		return err
	}
	_, err = modifier.signer.Sign(request, bytes.NewReader(requestContent),
		apiGatewaySvcNameForSignature, modifier.signatureRegion, time.Now().UTC())
	return err
}

func readRequestBody(requestBody io.ReadCloser) ([]byte, error) {
	if requestBody == nil {
		return nil, nil
	}
	return ioutil.ReadAll(requestBody)
}
