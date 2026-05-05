package executor

import (
	"errors"
	"org/api-core/internal/db"
)

type Executor struct{}

func NewExecutor() *Executor {
	return &Executor{}
}

func (e *Executor) ExecuteStep(step db.WorkflowStep) error {
	switch step.StepType {

	case "http_request":
		return e.executeHTTPRequest(step.Config)

	default:
		return errors.New("unsupported step type: " + step.StepType)
	}
}