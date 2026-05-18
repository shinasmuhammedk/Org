package handler

import (
	"net/http"

	usageService "org/api-core/internal/usage/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type UsageHandler struct {
	usageService *usageService.Service
}

func NewUsageHandler(usageService *usageService.Service) *UsageHandler {
	return &UsageHandler{
		usageService: usageService,
	}
}

func (h *UsageHandler) GetUsage(c *gin.Context) {
	userIDString := c.GetString("user_id")

	userID, err := uuid.Parse(userIDString)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "invalid user",
		})
		return
	}

	usage, err := h.usageService.GetCurrentMonthUsage(
		c.Request.Context(),
		userID,
	)

	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"month":         h.usageService.CurrentMonth(),
			"workflow_runs": 0,
		})
		return
	}

	runs := int32(0)
	if usage.WorkflowRuns.Valid {
		runs = usage.WorkflowRuns.Int32
	}

	c.JSON(http.StatusOK, gin.H{
		"month":         usage.Month,
		"workflow_runs": runs,
	})
}