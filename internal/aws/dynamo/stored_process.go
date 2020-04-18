package dynamo

import (
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/nordcloud/termination-detector/internal/process"
	"github.com/pkg/errors"
)

type storedProcess struct {
	ID               string        `json:"id"`
	State            process.State `json:"state"`
	StateDescription *string       `json:"state_description"`
}

func (proc storedProcess) dynamoItem() (map[string]*dynamodb.AttributeValue, error) {
	dynamoItem, err := dynamodbattribute.MarshalMap(proc)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to marshal stored process: %+v", proc)
	}
	return dynamoItem, nil
}

func newStoredProcess(proc process.Process) storedProcess {
	return storedProcess{
		ID:               proc.ID,
		State:            proc.State,
		StateDescription: proc.StateDescription,
	}
}
