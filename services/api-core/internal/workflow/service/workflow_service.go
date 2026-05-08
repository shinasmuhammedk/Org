package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

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

type SaveWorkflowEdgeRequest struct {
	Source string `json:"source"`
	Target string `json:"target"`
}

type SaveWorkflowStepsRequest struct {
	Steps []SaveWorkflowStepRequest `json:"steps"`
	Edges []SaveWorkflowEdgeRequest `json:"edges"`
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

func (s *WorkflowService) RunWorkflow(ctx context.Context, workflowID uuid.UUID, userID uuid.UUID) (uuid.UUID, error) {
	runID := uuid.New()

	_, err := s.repo.CreateWorkflowRun(ctx, db.CreateWorkflowRunParams{
		ID:         runID,
		WorkflowID: workflowID,
		UserID:     userID,
		Status:     "running",
	})
	if err != nil {
		return runID, err
	}

	failRun := func(execErr error) error {
		_ = s.repo.UpdateWorkflowRunStatus(ctx, db.UpdateWorkflowRunStatusParams{
			ID:     runID,
			Status: "failed",
			ErrorMessage: sql.NullString{
				String: execErr.Error(),
				Valid:  true,
			},
		})
		return execErr
	}

	steps, err := s.repo.ListWorkflowSteps(ctx, workflowID)
	if err != nil {
		return runID, failRun(err)
	}

	edges, err := s.repo.ListWorkflowEdgesForExecution(ctx, workflowID)
	if err != nil {
		return runID, failRun(err)
	}

	if len(steps) == 0 {
		return runID, failRun(errors.New("workflow has no steps"))
	}

	stepMap := make(map[uuid.UUID]db.WorkflowStep)
	incomingCount := make(map[uuid.UUID]int)
	nextSteps := make(map[uuid.UUID][]uuid.UUID)

	for _, step := range steps {
		stepMap[step.ID] = step
		incomingCount[step.ID] = 0
	}

	for _, edge := range edges {
		nextSteps[edge.SourceStepID] = append(nextSteps[edge.SourceStepID], edge.TargetStepID)
		incomingCount[edge.TargetStepID]++
	}

	var startStepID uuid.UUID
	foundStart := false

	for _, step := range steps {
		if incomingCount[step.ID] == 0 {
			startStepID = step.ID
			foundStart = true
			break
		}
	}

	if !foundStart {
		return runID, failRun(errors.New("no start node found"))
	}

	exec := executor.NewExecutor()
	visited := make(map[uuid.UUID]bool)

	var executeNode func(execCtx context.Context, stepID uuid.UUID) error

	executeNode = func(execCtx context.Context, stepID uuid.UUID) error {
		if visited[stepID] {
			return nil
		}

		step, ok := stepMap[stepID]
		if !ok {
			return errors.New("step not found in workflow")
		}

		visited[stepID] = true

		stepRunID := uuid.New()

		_, err := s.repo.CreateWorkflowStepRun(execCtx, db.CreateWorkflowStepRunParams{
			ID:             stepRunID,
			WorkflowRunID:  runID,
			WorkflowStepID: step.ID,
			Status:         "running",
			Input: pqtype.NullRawMessage{
				RawMessage: step.Config,
				Valid:      true,
			},
			Output: pqtype.NullRawMessage{
				Valid: false,
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
			_ = s.repo.UpdateWorkflowStepRunStatus(execCtx, db.UpdateWorkflowStepRunStatusParams{
				ID:     stepRunID,
				Status: "failed",
				Output: pqtype.NullRawMessage{
					RawMessage: output,
					Valid:      output != nil,
				},
				ErrorMessage: sql.NullString{
					String: err.Error(),
					Valid:  true,
				},
			})

			return err
		}

		err = s.repo.UpdateWorkflowStepRunStatus(execCtx, db.UpdateWorkflowStepRunStatusParams{
			ID:     stepRunID,
			Status: "success",
			Output: pqtype.NullRawMessage{
				RawMessage: output,
				Valid:      true,
			},
			ErrorMessage: sql.NullString{
				Valid: false,
			},
		})
		if err != nil {
			return err
		}

		for _, nextStepID := range nextSteps[stepID] {
			if err := executeNode(execCtx,nextStepID); err != nil {
				return err
			}
		}

		return nil
	}

	go func() {

		bgCtx := context.Background()

		if err := executeNode(bgCtx, startStepID); err != nil {
			_ = failRun(err)
			return
		}

		_ = s.repo.UpdateWorkflowRunStatus(
			context.Background(),
			db.UpdateWorkflowRunStatusParams{
				ID:     runID,
				Status: "success",
				ErrorMessage: sql.NullString{
					Valid: false,
				},
			},
		)

	}()

	return runID, nil
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
	edges []SaveWorkflowEdgeRequest,
) error {
	_, err := s.repo.GetWorkflowByID(ctx, db.GetWorkflowByIDParams{
		ID:     workflowID,
		UserID: userID,
	})
	if err != nil {
		return err
	}

	err = s.repo.DeleteWorkflowEdges(ctx, workflowID)
	if err != nil {
		return err
	}

	err = s.repo.DeleteWorkflowSteps(ctx, workflowID)
	if err != nil {
		return err
	}

	nodeIDToStepID := make(map[string]uuid.UUID)

	for _, step := range steps {
		stepID := uuid.New()

		_, err := s.repo.CreateWorkflowStep(ctx, db.CreateWorkflowStepParams{
			ID:             stepID,
			WorkflowID:     workflowID,
			FrontendNodeID: step.FrontendNodeID,
			StepOrder:      int32(step.StepOrder),
			StepType:       step.StepType,
			Config:         step.Config,
		})
		if err != nil {
			return err
		}

		nodeIDToStepID[step.FrontendNodeID] = stepID
	}

	for _, edge := range edges {
		sourceStepID, ok := nodeIDToStepID[edge.Source]
		if !ok {
			return fmt.Errorf("source node not found: %s", edge.Source)
		}

		targetStepID, ok := nodeIDToStepID[edge.Target]
		if !ok {
			return fmt.Errorf("target node not found: %s", edge.Target)
		}

		_, err := s.repo.CreateWorkflowEdge(ctx, db.CreateWorkflowEdgeParams{
			ID:           uuid.New(),
			WorkflowID:   workflowID,
			SourceStepID: sourceStepID,
			TargetStepID: targetStepID,
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

func (s *WorkflowService) GetWorkflowEdges(
	ctx context.Context,
	workflowID uuid.UUID,
	userID uuid.UUID,
) ([]db.ListWorkflowEdgesRow, error) {

	_, err := s.repo.GetWorkflowByID(ctx, db.GetWorkflowByIDParams{
		ID:     workflowID,
		UserID: userID,
	})
	if err != nil {
		return nil, err
	}

	return s.repo.ListWorkflowEdges(ctx, workflowID)
}
