package repository

import (
	"context"

	"github.com/google/uuid"
	"org/api-core/internal/db"
)

type WorkflowRepository interface {
	// Workflow
	CreateWorkflow(ctx context.Context, arg db.CreateWorkflowParams) (db.Workflow, error)
	ListWorkflowByUser(ctx context.Context, userID uuid.UUID) ([]db.Workflow, error)
	GetWorkflowByID(ctx context.Context, arg db.GetWorkflowByIDParams) (db.Workflow, error)
	DeleteWorkflow(ctx context.Context, arg db.DeleteWorkflowParams) error

	// Steps
	CreateWorkflowStep(ctx context.Context, arg db.CreateWorkflowStepParams) (db.WorkflowStep, error)
	ListWorkflowSteps(ctx context.Context, workflowID uuid.UUID) ([]db.WorkflowStep, error)

	// Runs
	CreateWorkflowRun(ctx context.Context, arg db.CreateWorkflowRunParams) (db.WorkflowRun, error)
	UpdateWorkflowRunStatus(ctx context.Context, arg db.UpdateWorkflowRunStatusParams) error
	ListWorkflowRuns(ctx context.Context, arg db.ListWorkflowRunsParams) ([]db.WorkflowRun, error)

	CreateWorkflowStepRun(ctx context.Context, arg db.CreateWorkflowStepRunParams) (db.WorkflowStepRun, error)
	ListWorkflowStepRuns(ctx context.Context, workflowRunID uuid.UUID) ([]db.WorkflowStepRun, error)
    UpdateWorkflowStepRunStatus(ctx context.Context, arg db.UpdateWorkflowStepRunStatusParams) error
}
