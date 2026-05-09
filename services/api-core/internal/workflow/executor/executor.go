package executor

import (
	"errors"

	"org/api-core/internal/db"
)

type Executor struct{}

func NewExecutor() *Executor {
	return &Executor{}
}

func (e *Executor) ExecuteStep(
	step db.WorkflowStep,
	input []byte,
) ([]byte, error) {

	switch step.StepType {

	case "webhook_trigger":

		if len(input) > 0 {
			return input, nil
		}

		return step.Config, nil

	case "http_request":
		return e.executeHTTPRequest(step.Config, input)

	case "condition":
		return e.executeCondition(step.Config, input)

	case "delay":
		return e.executeDelay(step.Config)

	default:
		return nil, errors.New(
			"unsupported step type: " + step.StepType,
		)
	}
}
