package executor

import (
	"errors"

	"org/api-core/internal/db"
)

type Executor struct{}

func NewExecutor() *Executor {
	return &Executor{}
}

func (e *Executor) ExecuteStep(step db.WorkflowStep) ([]byte, error) {
	switch step.StepType {
	case "webhook_trigger":
		return step.Config, nil

	case "http_request":
		return e.executeHTTPRequest(step.Config)

	default:
		return nil, errors.New("unsupported step type: " + step.StepType)
	}
}