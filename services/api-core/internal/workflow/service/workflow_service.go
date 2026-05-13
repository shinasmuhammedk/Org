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
	PositionX      float64         `json:"position_x"`
	PositionY      float64         `json:"position_y"`
}

type SaveWorkflowEdgeRequest struct {
	Source          string `json:"source"`
	Target          string `json:"target"`
	ConditionBranch string `json:"condition_branch"`
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

// func (s *WorkflowService) RunWorkflowWithInput(ctx context.Context, workflowID uuid.UUID, userID uuid.UUID) (uuid.UUID, error) {
// 	runID := uuid.New()

// 	_, err := s.repo.CreateWorkflowRun(ctx, db.CreateWorkflowRunParams{
// 		ID:         runID,
// 		WorkflowID: workflowID,
// 		UserID:     userID,
// 		Status:     "running",
// 	})
// 	if err != nil {
// 		return runID, err
// 	}

// 	failRun := func(execErr error) error {
// 		_ = s.repo.UpdateWorkflowRunStatus(ctx, db.UpdateWorkflowRunStatusParams{
// 			ID:     runID,
// 			Status: "failed",
// 			ErrorMessage: sql.NullString{
// 				String: execErr.Error(),
// 				Valid:  true,
// 			},
// 		})
// 		return execErr
// 	}

// 	steps, err := s.repo.ListWorkflowSteps(ctx, workflowID)
// 	if err != nil {
// 		return runID, failRun(err)
// 	}

// 	edges, err := s.repo.ListWorkflowEdgesForExecution(ctx, workflowID)
// 	if err != nil {
// 		return runID, failRun(err)
// 	}

// 	if len(steps) == 0 {
// 		return runID, failRun(errors.New("workflow has no steps"))
// 	}

// 	stepMap := make(map[uuid.UUID]db.WorkflowStep)
// 	incomingCount := make(map[uuid.UUID]int)
// 	nextSteps := make(map[uuid.UUID][]uuid.UUID)

// 	for _, step := range steps {
// 		stepMap[step.ID] = step
// 		incomingCount[step.ID] = 0
// 	}

// 	for _, edge := range edges {
// 		nextSteps[edge.SourceStepID] = append(nextSteps[edge.SourceStepID], edge.TargetStepID)
// 		incomingCount[edge.TargetStepID]++
// 	}

// 	var startStepID uuid.UUID
// 	foundStart := false

// 	for _, step := range steps {
// 		if incomingCount[step.ID] == 0 {
// 			startStepID = step.ID
// 			foundStart = true
// 			break
// 		}
// 	}

// 	if !foundStart {
// 		return runID, failRun(errors.New("no start node found"))
// 	}

// 	exec := executor.NewExecutor()
// 	visited := make(map[uuid.UUID]bool)

// 	var executeNode func(execCtx context.Context, stepID uuid.UUID) error

// 	executeNode = func(execCtx context.Context, stepID uuid.UUID) error {
// 		if visited[stepID] {
// 			return nil
// 		}

// 		step, ok := stepMap[stepID]
// 		if !ok {
// 			return errors.New("step not found in workflow")
// 		}

// 		visited[stepID] = true

// 		stepRunID := uuid.New()

// 		_, err := s.repo.CreateWorkflowStepRun(execCtx, db.CreateWorkflowStepRunParams{
// 			ID:             stepRunID,
// 			WorkflowRunID:  runID,
// 			WorkflowStepID: step.ID,
// 			Status:         "running",
// 			Input: pqtype.NullRawMessage{
// 				RawMessage: step.Config,
// 				Valid:      true,
// 			},
// 			Output: pqtype.NullRawMessage{
// 				Valid: false,
// 			},
// 			ErrorMessage: sql.NullString{
// 				Valid: false,
// 			},
// 		})
// 		if err != nil {
// 			return err
// 		}

// 		output, err := exec.ExecuteStep(step)
// 		if err != nil {
// 			_ = s.repo.UpdateWorkflowStepRunStatus(execCtx, db.UpdateWorkflowStepRunStatusParams{
// 				ID:     stepRunID,
// 				Status: "failed",
// 				Output: pqtype.NullRawMessage{
// 					RawMessage: output,
// 					Valid:      output != nil,
// 				},
// 				ErrorMessage: sql.NullString{
// 					String: err.Error(),
// 					Valid:  true,
// 				},
// 			})

// 			return err
// 		}

// 		err = s.repo.UpdateWorkflowStepRunStatus(execCtx, db.UpdateWorkflowStepRunStatusParams{
// 			ID:     stepRunID,
// 			Status: "success",
// 			Output: pqtype.NullRawMessage{
// 				RawMessage: output,
// 				Valid:      true,
// 			},
// 			ErrorMessage: sql.NullString{
// 				Valid: false,
// 			},
// 		})
// 		if err != nil {
// 			return err
// 		}

// 		for _, nextStepID := range nextSteps[stepID] {
// 			if err := executeNode(execCtx, nextStepID); err != nil {
// 				return err
// 			}
// 		}

// 		return nil
// 	}

// 	go func() {

// 		bgCtx := context.Background()

// 		if err := executeNode(bgCtx, startStepID); err != nil {
// 			_ = failRun(err)
// 			return
// 		}

// 		_ = s.repo.UpdateWorkflowRunStatus(
// 			context.Background(),
// 			db.UpdateWorkflowRunStatusParams{
// 				ID:     runID,
// 				Status: "success",
// 				ErrorMessage: sql.NullString{
// 					Valid: false,
// 				},
// 			},
// 		)

// 	}()

// 	return runID, nil
// }

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

	err = s.repo.DeleteWebhookTriggersByWorkflow(ctx, workflowID)
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

		config := step.Config

		if step.StepType == "webhook_trigger" {
			webhookURLID := uuid.NewString()

			var configMap map[string]interface{}

			if len(config) > 0 {
				_ = json.Unmarshal(config, &configMap)
			}

			if configMap == nil {
				configMap = make(map[string]interface{})
			}

			configMap["webhook_url_id"] = webhookURLID
			configMap["webhook_url"] = "http://localhost:8080/webhooks/" + webhookURLID

			updatedConfig, err := json.Marshal(configMap)
			if err != nil {
				return err
			}

			config = updatedConfig

			_, err = s.repo.CreateWebhookTrigger(ctx, db.CreateWebhookTriggerParams{
				ID:             uuid.New(),
				WorkflowID:     workflowID,
				UserID:         userID,
				WebhookUrlID:   webhookURLID,
				FrontendNodeID: step.FrontendNodeID,
			})
			if err != nil {
				return err
			}
		}

		_, err := s.repo.CreateWorkflowStep(ctx, db.CreateWorkflowStepParams{
			ID:             stepID,
			WorkflowID:     workflowID,
			FrontendNodeID: step.FrontendNodeID,
			StepOrder:      int32(step.StepOrder),
			StepType:       step.StepType,
			Config:         config,
			PositionX:      step.PositionX,
			PositionY:      step.PositionY,
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
			ConditionBranch: sql.NullString{
				String: edge.ConditionBranch,
				Valid:  edge.ConditionBranch != "",
			},
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

func (s *WorkflowService) RunWorkflowFromWebhook(
	ctx context.Context,
	webhookID string,
	payload map[string]interface{},
) (uuid.UUID, error) {
	trigger, err := s.repo.GetWebhookTriggerByURLID(ctx, webhookID)
	if err != nil {
		return uuid.Nil, err
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return uuid.Nil, err
	}

	return s.RunWorkflowWithInput(
		ctx,
		trigger.WorkflowID,
		trigger.UserID,
		payloadBytes,
	)
}

func (s *WorkflowService) RunWorkflow(
	ctx context.Context,
	workflowID uuid.UUID,
	userID uuid.UUID,
) (uuid.UUID, error) {
	return s.RunWorkflowWithInput(ctx, workflowID, userID, nil)
}

func (s *WorkflowService) RunWorkflowWithInput(
	ctx context.Context,
	workflowID uuid.UUID,
	userID uuid.UUID,
	initialInput []byte,
) (uuid.UUID, error) {
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

	failRun := func(execCtx context.Context, execErr error) error {
		_ = s.repo.UpdateWorkflowRunStatus(execCtx, db.UpdateWorkflowRunStatusParams{
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
		return runID, failRun(ctx, err)
	}

	edges, err := s.repo.ListWorkflowEdgesForExecution(ctx, workflowID)
	if err != nil {
		return runID, failRun(ctx, err)
	}

	if len(steps) == 0 {
		return runID, failRun(ctx, errors.New("workflow has no steps"))
	}

	stepMap := make(map[uuid.UUID]db.WorkflowStep)
	incomingCount := make(map[uuid.UUID]int)

	type NextEdge struct {
		TargetStepID    uuid.UUID
		ConditionBranch string
	}

	nextSteps := make(map[uuid.UUID][]NextEdge)

	for _, step := range steps {
		stepMap[step.ID] = step
		incomingCount[step.ID] = 0
	}

	for _, edge := range edges {

		branch := ""

		if edge.ConditionBranch.Valid {
			branch = edge.ConditionBranch.String
		}

		nextSteps[edge.SourceStepID] = append(
			nextSteps[edge.SourceStepID],
			NextEdge{
				TargetStepID:    edge.TargetStepID,
				ConditionBranch: branch,
			},
		)

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
		return runID, failRun(ctx, errors.New("no start node found"))
	}

	exec := executor.NewExecutor()
	visited := make(map[uuid.UUID]bool)
	nodeInputs := make(map[uuid.UUID][]byte)

	if len(initialInput) > 0 {
		nodeInputs[startStepID] = initialInput
	}

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
		input := nodeInputs[stepID]

		if len(input) == 0 {
			input = step.Config
		}

		_, err := s.repo.CreateWorkflowStepRun(execCtx, db.CreateWorkflowStepRunParams{
			ID:             stepRunID,
			WorkflowRunID:  runID,
			WorkflowStepID: step.ID,
			Status:         "running",
			Input: pqtype.NullRawMessage{
				RawMessage: input,
				Valid:      len(input) > 0,
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

		output, err := exec.ExecuteStep(step, nodeInputs[stepID])
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
				Valid:      output != nil,
			},
			ErrorMessage: sql.NullString{
				Valid: false,
			},
		})
		if err != nil {
			return err
		}

		fmt.Println("===================================")
		fmt.Println("CURRENT STEP:", step.StepType)
		fmt.Println("CURRENT STEP ID:", stepID)
		fmt.Println("NEXT EDGES:", nextSteps[stepID])

		for _, edge := range nextSteps[stepID] {

			fmt.Println("EDGE TARGET:", edge.TargetStepID)
			fmt.Println("EDGE BRANCH:", edge.ConditionBranch)

			if step.StepType == "condition" {

				var conditionResult struct {
					Result bool `json:"result"`
				}

				if err := json.Unmarshal(output, &conditionResult); err != nil {
					return err
				}

				fmt.Println("CONDITION RESULT:", conditionResult.Result)

				if conditionResult.Result && edge.ConditionBranch != "true" {
					fmt.Println("SKIPPING FALSE EDGE")
					continue
				}

				if !conditionResult.Result && edge.ConditionBranch != "false" {
					fmt.Println("SKIPPING TRUE EDGE")
					continue
				}
			}

			fmt.Println("EXECUTING NEXT NODE:", edge.TargetStepID)

			if step.StepType == "condition" {
				nodeInputs[edge.TargetStepID] = nodeInputs[stepID]
			} else {
				nodeInputs[edge.TargetStepID] = output
			}

			if err := executeNode(execCtx, edge.TargetStepID); err != nil {
				return err
			}
		}

		return nil
	}

	go func() {
		bgCtx := context.Background()

		if err := executeNode(bgCtx, startStepID); err != nil {
			_ = failRun(bgCtx, err)
			return
		}

		_ = s.repo.UpdateWorkflowRunStatus(bgCtx, db.UpdateWorkflowRunStatusParams{
			ID:     runID,
			Status: "success",
			ErrorMessage: sql.NullString{
				Valid: false,
			},
		})
	}()

	return runID, nil
}
