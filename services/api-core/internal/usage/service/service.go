package service

import (
	"context"
	"time"

	usageRepo "org/api-core/internal/usage/repository"

	"github.com/google/uuid"
    "org/api-core/internal/db"
)

type Service struct {
	repo usageRepo.Repository
}

func NewService(repo usageRepo.Repository) *Service {
	return &Service{
		repo: repo,
	}
}

func currentMonth() string {
	return time.Now().Format("2006-01")
}

func (s *Service) CanRunWorkflow(
	ctx context.Context,
	userID uuid.UUID,
	plan string,
) (bool, string, error) {
	month := currentMonth()

	usage, err := s.repo.GetMonthlyUsage(ctx, userID, month)
	if err != nil {
		// no usage record yet, allow
		return true, "", nil
	}

	limit := 100

	if plan != "free" {
		limit = 10000
	}

	runs := int32(0)

	if usage.WorkflowRuns.Valid {
		runs = usage.WorkflowRuns.Int32
	}

	if runs >= int32(limit) {
		return false, "monthly workflow run limit reached", nil
	}

	return true, "", nil
}

func (s *Service) IncrementWorkflowRun(
	ctx context.Context,
	userID uuid.UUID,
) error {
	_, err := s.repo.IncrementWorkflowUsage(
		ctx,
		userID,
		currentMonth(),
	)

	return err
}


func (s *Service) CurrentMonth() string {
	return currentMonth()
}

func (s *Service) GetCurrentMonthUsage(
	ctx context.Context,
	userID uuid.UUID,
) (db.WorkflowUsage, error) {
	return s.repo.GetMonthlyUsage(
		ctx,
		userID,
		currentMonth(),
	)
}