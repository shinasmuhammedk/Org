package handler

import (
	"encoding/json"
	"log"

	"Org/utils/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"org/api-core/internal/billing"
	usageService "org/api-core/internal/usage/service"
	"org/api-core/internal/workflow/service"
)

type WorkflowHandler struct {
	workflowService *service.WorkflowService
	usageService    *usageService.Service
}

func NewWorkflowHandler(workflowService *service.WorkflowService, usageService *usageService.Service) *WorkflowHandler {
	return &WorkflowHandler{
		workflowService: workflowService,
		usageService:    usageService,
	}
}

type SaveWorkflowStepsRequest struct {
	Steps []service.SaveWorkflowStepRequest `json:"steps"`
	Edges []service.SaveWorkflowEdgeRequest `json:"edges"`
}

type SaveWorkflowRequest struct {
	Name string `json:"name"`
}

type UpdateWorkflowScheduleRequest struct {
	ScheduleEnabled bool   `json:"schedule_enabled"`
	ScheduleType    string `json:"schedule_type"`
	ScheduleValue   string `json:"schedule_value"`
}

// CREATE WORKFLOW
func (h *WorkflowHandler) CreateWorkflow(c *gin.Context) {
	var body struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "invalid input", err.Error())
		return
	}

	if body.Name == "" {
		response.BadRequest(c, "workflow name is required", nil)
		return
	}

	userIDValue, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "unauthorized")
		return
	}

	userID, err := uuid.Parse(userIDValue.(string))
	if err != nil {
		response.BadRequest(c, "invalid user id", err.Error())
		return
	}

	workflow, err := h.workflowService.CreateWorkflow(
		c.Request.Context(),
		userID,
		body.Name,
		body.Description,
	)

	if err != nil {
		response.InternalServerError(c, "failed to create workflow", err.Error())
		return
	}

	response.Created(c, "workflow created successfully", workflow)
}

// LIST WORKFLOWS
func (h *WorkflowHandler) ListWorkflows(c *gin.Context) {
	userIDValue, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "unauthorized")
		return
	}

	userID, err := uuid.Parse(userIDValue.(string))
	if err != nil {
		response.BadRequest(c, "invalid user id", err.Error())
		return
	}

	workflows, err := h.workflowService.ListWorkflow(c.Request.Context(), userID)
	if err != nil {
		response.InternalServerError(c, "failed to fetch workflows", err.Error())
		return
	}

	response.OK(c, "workflows fetched successfully", workflows)
}

// DELETE WORKFLOW
func (h *WorkflowHandler) DeleteWorkflow(c *gin.Context) {
	userIDValue, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "unauthorized")
		return
	}

	userID, err := uuid.Parse(userIDValue.(string))
	if err != nil {
		response.BadRequest(c, "invalid user id", err.Error())
		return
	}

	workflowID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid workflow id", err.Error())
		return
	}

	err = h.workflowService.DeleteWorkflow(c.Request.Context(), userID, workflowID)
	if err != nil {
		response.InternalServerError(c, "failed to delete workflow", err.Error())
		return
	}

	response.OK(c, "workflow deleted successfully", nil)
}

// CREATE SINGLE STEP
func (h *WorkflowHandler) CreateStep(c *gin.Context) {
	var body struct {
		StepOrder int32                  `json:"step_order"`
		StepType  string                 `json:"step_type"`
		Config    map[string]interface{} `json:"config"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "invalid input", err.Error())
		return
	}

	if body.StepType == "" {
		response.BadRequest(c, "step type is required", nil)
		return
	}

	workflowID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid workflow id", err.Error())
		return
	}

	configBytes, err := json.Marshal(body.Config)
	if err != nil {
		response.BadRequest(c, "invalid config", err.Error())
		return
	}

	step, err := h.workflowService.CreateStep(
		c.Request.Context(),
		workflowID,
		body.StepOrder,
		body.StepType,
		configBytes,
	)

	if err != nil {
		response.InternalServerError(c, "failed to create step", err.Error())
		return
	}

	response.Created(c, "workflow step created successfully", step)
}

// LIST STEPS
func (h *WorkflowHandler) ListSteps(c *gin.Context) {
	workflowID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid workflow id", err.Error())
		return
	}

	steps, err := h.workflowService.ListSteps(c.Request.Context(), workflowID)
	if err != nil {
		response.InternalServerError(c, "failed to fetch steps", err.Error())
		return
	}

	response.OK(c, "workflow steps fetched successfully", steps)
}

// SAVE CANVAS STEPS
func (h *WorkflowHandler) SaveWorkflowSteps(c *gin.Context) {
	workflowID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid workflow id", err.Error())
		return
	}

	userIDString := c.MustGet("user_id").(string)

	userID, err := uuid.Parse(userIDString)
	if err != nil {
		response.Unauthorized(c, "invalid user")
		return
	}

	var body SaveWorkflowStepsRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "invalid request body", err.Error())
		return
	}

	err = h.workflowService.SaveWorkflowSteps(
		c.Request.Context(),
		workflowID,
		userID,
		body.Steps,
		body.Edges,
	)

	if err != nil {
		response.InternalServerError(c, "failed to save workflow steps", err.Error())
		return
	}

	response.OK(c, "workflow steps saved successfully", nil)
}

// GET CANVAS STEPS
func (h *WorkflowHandler) GetWorkflowSteps(c *gin.Context) {
	workflowID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid workflow id", err.Error())
		return
	}

	userIDString := c.MustGet("user_id").(string)

	userID, err := uuid.Parse(userIDString)
	if err != nil {
		response.Unauthorized(c, "invalid user")
		return
	}

	steps, err := h.workflowService.GetWorkflowSteps(
		c.Request.Context(),
		workflowID,
		userID,
	)

	if err != nil {
		response.InternalServerError(c, "failed to fetch workflow steps", err.Error())
		return
	}

	response.OK(c, "workflow steps fetched successfully", steps)
}

// RUN WORKFLOW
func (h *WorkflowHandler) RunWorkflow(c *gin.Context) {
	workflowID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid workflow id", err.Error())
		return
	}

	userIDString := c.MustGet("user_id").(string)

	userID, err := uuid.Parse(userIDString)
	if err != nil {
		response.Unauthorized(c, "invalid user")
		return
	}

	// CHECK SUBSCRIPTION
	plan, status, err := billing.GetUserSubscription(userIDString)
	if err != nil {
		response.InternalServerError(c, "failed to check subscription", err.Error())
		return
	}

	if status != "active" {
		response.Forbidden(c, "subscription inactive")
		return
	}

	if plan == "free" {
		response.Forbidden(c, "upgrade required to run workflows")
		return
	}

	// CHECK USAGE LIMIT BEFORE RUNNING
	allowed, message, err := h.usageService.CanRunWorkflow(
		c.Request.Context(),
		userID,
		plan,
	)
	if err != nil {
		response.InternalServerError(c, "failed to check usage", err.Error())
		return
	}

	if !allowed {
		response.Forbidden(c, message)
		return
	}

	runID, err := h.workflowService.RunWorkflow(
		c.Request.Context(),
		workflowID,
		userID,
	)
	if err != nil {
		response.InternalServerError(c, "workflow execution failed", err.Error())
		return
	}

	// INCREMENT USAGE AFTER SUCCESSFUL RUN
	if err := h.usageService.IncrementWorkflowRun(c.Request.Context(), userID); err != nil {
		// do not block response, just log
		log.Println("failed to increment workflow usage:", err)
	}

	response.OK(c, "workflow executed successfully", gin.H{
		"run_id": runID,
	})
}

// LIST WORKFLOW RUNS
func (h *WorkflowHandler) ListWorkflowRuns(c *gin.Context) {
	workflowID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid workflow id", err.Error())
		return
	}

	userIDString := c.MustGet("user_id").(string)

	userID, err := uuid.Parse(userIDString)
	if err != nil {
		response.Unauthorized(c, "invalid user")
		return
	}

	runs, err := h.workflowService.ListWorkflowRuns(c.Request.Context(), workflowID, userID)
	if err != nil {
		response.InternalServerError(c, "failed to fetch workflow runs", err.Error())
		return
	}

	response.OK(c, "workflow runs fetched successfully", runs)
}

// LIST STEP RUN LOGS
func (h *WorkflowHandler) ListWorkflowStepRuns(c *gin.Context) {
	runID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid workflow run id", err.Error())
		return
	}

	steps, err := h.workflowService.ListWorkflowStepRuns(c.Request.Context(), runID)
	if err != nil {
		response.InternalServerError(c, "failed to fetch step logs", err.Error())
		return
	}

	response.OK(c, "workflow step logs fetched successfully", steps)
}

func (h *WorkflowHandler) GetWorkflowEdges(c *gin.Context) {
	workflowID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid workflow id", err.Error())
		return
	}

	userIDString := c.MustGet("user_id").(string)

	userID, err := uuid.Parse(userIDString)
	if err != nil {
		response.Unauthorized(c, "invalid user")
		return
	}

	edges, err := h.workflowService.GetWorkflowEdges(
		c.Request.Context(),
		workflowID,
		userID,
	)
	if err != nil {
		response.InternalServerError(c, "failed to fetch workflow edges", err.Error())
		return
	}

	response.OK(c, "workflow edges fetched successfully", edges)
}

func (h *WorkflowHandler) HandleWebhookTrigger(c *gin.Context) {
	webhookId := c.Param("webhookID")

	var payload map[string]interface{}

	if err := c.ShouldBindJSON(&payload); err != nil {
		response.BadRequest(c, "invalid webhook payload", err.Error())
		return
	}

	runID, err := h.workflowService.RunWorkflowFromWebhook(c.Request.Context(), webhookId, payload)
	if err != nil {
		response.InternalServerError(c, "webhook execution failed", err.Error())
		return
	}

	response.OK(c, "webhook recieved", gin.H{
		"run_id": runID,
	})
}

func (h *WorkflowHandler) UpdateWorkflowSchedule(c *gin.Context) {
	workflowIDParam := c.Param("id")

	workflowID, err := uuid.Parse(workflowIDParam)
	if err != nil {
		response.BadRequest(c, "invalid workflow id", nil)
		return
	}

	userIDValue, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "unauthorized")
		return
	}

	userIDString, ok := userIDValue.(string)
	if !ok {
		response.Unauthorized(c, "invalid user")
		return
	}

	userID, err := uuid.Parse(userIDString)
	if err != nil {
		response.Unauthorized(c, "invalid user id")
		return
	}

	var req service.UpdateWorkflowScheduleRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request body", nil)
		return
	}

	err = h.workflowService.UpdateWorkflowSchedule(
		c.Request.Context(),
		workflowID,
		userID,
		req,
	)

	if err != nil {
		response.InternalServerError(c, "failed to update workflow schedule", err.Error())
		return
	}

	response.OK(c, "workflow schedule updated", nil)
}

func (h *WorkflowHandler) GetWorkflowSchedule(c *gin.Context) {
	workflowID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid workflow id", nil)
		return
	}

	userIDValue, _ := c.Get("user_id")
	userIDString := userIDValue.(string)

	userID, err := uuid.Parse(userIDString)
	if err != nil {
		response.BadRequest(c, "invalid user id", nil)
		return
	}

	schedule, err := h.workflowService.GetWorkflowSchedule(
		c.Request.Context(),
		workflowID,
		userID,
	)

	if err != nil {
		response.InternalServerError(c, "failed to get workflow schedule", err.Error())
		return
	}

	response.OK(c, "workflow schedule fetched", schedule)
}
