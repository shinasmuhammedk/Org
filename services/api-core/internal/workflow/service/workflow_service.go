package service

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/sqlc-dev/pqtype"

	"org/api-core/internal/db"
	"org/api-core/internal/workflow/executor"
	"org/api-core/internal/workflow/repository"
)

type WorkflowService struct {
	repo repository.WorkflowRepository
}

func NewWorkflowService(repo repository.WorkflowRepository) *WorkflowService {
	return &WorkflowService{repo: repo}
}

type SaveWorkflowStepRequest struct {
	FrontendNodeID string          `json:"frontend_node_id"`
	StepOrder      int             `json:"step_order"`
	StepType       string          `json:"step_type"`
	Config         json.RawMessage `json:"config"`
}

type SaveWorkflowStepsRequest struct {
	Steps []SaveWorkflowStepRequest `json:"steps"`
}

//
// 🔹 WORKFLOW METHODS
//

func (s *WorkflowService) CreateWorkflow(ctx context.Context, userID uuid.UUID, name string, description string) (db.Workflow, error) {

	return s.repo.CreateWorkflow(ctx, db.CreateWorkflowParams{
		ID:          uuid.New(),
		UserID:      userID,
		Name:        name,
		Description: description,
		TriggerType: "manual",
		IsActive:    true,
	})
}

func (s *WorkflowService) ListWorkflow(ctx context.Context, userID uuid.UUID) ([]db.Workflow, error) {

	return s.repo.ListWorkflowByUser(ctx, userID)
}

func (s *WorkflowService) DeleteWorkflow(ctx context.Context, userID uuid.UUID, workflowID uuid.UUID) error {

	return s.repo.DeleteWorkflow(ctx, db.DeleteWorkflowParams{
		ID:     workflowID,
		UserID: userID,
	})
}

//
// 🔹 STEP METHODS
//

func (s *WorkflowService) CreateStep(ctx context.Context, workflowID uuid.UUID, stepOrder int32, stepType string, config []byte) (db.WorkflowStep, error) {

	return s.repo.CreateWorkflowStep(ctx, db.CreateWorkflowStepParams{
		ID:         uuid.New(),
		WorkflowID: workflowID,
		StepOrder:  stepOrder,
		StepType:   stepType,
		Config:     config,
	})
}

func (s *WorkflowService) ListSteps(
	ctx context.Context,
	workflowID uuid.UUID,
) ([]db.WorkflowStep, error) {

	return s.repo.ListWorkflowSteps(ctx, workflowID)
}

func (s *WorkflowService) RunWorkflow(ctx context.Context, workflowID uuid.UUID, userID uuid.UUID) error {
	runID := uuid.New()

	_, err := s.repo.CreateWorkflowRun(ctx, db.CreateWorkflowRunParams{
		ID:         runID,
		WorkflowID: workflowID,
		UserID:     userID,
		Status:     "running",
	})
	if err != nil {
		return err
	}

	steps, err := s.repo.ListWorkflowSteps(ctx, workflowID)
	if err != nil {
		errMsg := err.Error()

		_ = s.repo.UpdateWorkflowRunStatus(ctx, db.UpdateWorkflowRunStatusParams{
			ID:     runID,
			Status: "failed",
			ErrorMessage: sql.NullString{
				String: errMsg,
				Valid:  true,
			},
		})

		return err
	}

	exec := executor.NewExecutor()

	for _, step := range steps {
		stepRunID := uuid.New()

		_, err := s.repo.CreateWorkflowStepRun(ctx, db.CreateWorkflowStepRunParams{
			ID:             stepRunID,
			WorkflowRunID:  runID,
			WorkflowStepID: step.ID,
			Status:         "running",
			Input: pqtype.NullRawMessage{
				RawMessage: step.Config,
				Valid:      true,
			},
			Output: pqtype.NullRawMessage{
				RawMessage: step.Config,
				Valid:      true,
			},
			ErrorMessage: sql.NullString{
				Valid: false,
			},
		})
		if err != nil {
			return err
		}

		output, err := exec.ExecuteStep(step)
		if err != nil {
			errMsg := err.Error()

			_ = s.repo.UpdateWorkflowStepRunStatus(ctx, db.UpdateWorkflowStepRunStatusParams{
				ID:     stepRunID,
				Status: "failed",
				Output: pqtype.NullRawMessage{
					RawMessage: output,
					Valid:      true,
				},
				ErrorMessage: sql.NullString{
					String: errMsg,
					Valid:  true,
				},
			})

			_ = s.repo.UpdateWorkflowRunStatus(ctx, db.UpdateWorkflowRunStatusParams{
				ID:     runID,
				Status: "failed",
				ErrorMessage: sql.NullString{
					String: errMsg,
					Valid:  true,
				},
			})

			return err
		}

		err = s.repo.UpdateWorkflowStepRunStatus(ctx, db.UpdateWorkflowStepRunStatusParams{
			ID:     stepRunID,
			Status: "success",
			Output: pqtype.NullRawMessage{
				RawMessage: step.Config,
				Valid:      true,
			},
			ErrorMessage: sql.NullString{
				Valid: false,
			},
		})
		if err != nil {
			return err
		}
	}

	err = s.repo.UpdateWorkflowRunStatus(ctx, db.UpdateWorkflowRunStatusParams{
		ID:     runID,
		Status: "success",
		ErrorMessage: sql.NullString{
			Valid: false,
		},
	})
	if err != nil {
		return err
	}

	return nil
}

func (s *WorkflowService) ListWorkflowRuns(ctx context.Context, workflowID, userID uuid.UUID) ([]db.WorkflowRun, error) {
	return s.repo.ListWorkflowRuns(ctx, db.ListWorkflowRunsParams{
		WorkflowID: workflowID,
		UserID:     userID,
	})
}

func (s *WorkflowService) ListWorkflowStepRuns(ctx context.Context, workflowRunID uuid.UUID) ([]db.WorkflowStepRun, error) {
	return s.repo.ListWorkflowStepRuns(ctx, workflowRunID)
}

func (s *WorkflowService) SaveWorkflowSteps(
	ctx context.Context,
	workflowID uuid.UUID,
	userID uuid.UUID,
	steps []SaveWorkflowStepRequest,
) error {
	_, err := s.repo.GetWorkflowByID(ctx, db.GetWorkflowByIDParams{
		ID:     workflowID,
		UserID: userID,
	})
	if err != nil {
		return err
	}

	err = s.repo.DeleteWorkflowSteps(ctx, workflowID)
	if err != nil {
		return err
	}

	for _, step := range steps {
		_, err := s.repo.CreateWorkflowStep(ctx, db.CreateWorkflowStepParams{
			ID:             uuid.New(),
			WorkflowID:     workflowID,
			FrontendNodeID: step.FrontendNodeID,
			StepOrder:      int32(step.StepOrder),
			StepType:       step.StepType,
			Config:         step.Config,
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *WorkflowService) GetWorkflowSteps(
	ctx context.Context,
	workflowID uuid.UUID,
	userID uuid.UUID,
) ([]db.WorkflowStep, error) {
	_, err := s.repo.GetWorkflowByID(ctx, db.GetWorkflowByIDParams{
		ID:     workflowID,
		UserID: userID,
	})
	if err != nil {
		return nil, err
	}

	return s.repo.ListWorkflowSteps(ctx, workflowID)
}
