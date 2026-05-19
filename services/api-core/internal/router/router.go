package router

import (
	"context"

	"github.com/gin-gonic/gin"

	"org/api-core/internal/auth/handler"
	authHandlerPkg "org/api-core/internal/auth/handler"
	"org/api-core/internal/auth/middleware"
	authRepository "org/api-core/internal/auth/repository"
	authServicePkg "org/api-core/internal/auth/service"
	tokenstore "org/api-core/internal/auth/tokenstore"
	"org/api-core/internal/worker"

	"org/api-core/internal/db"

	oauthHandlerPkg "org/api-core/internal/oauth/handler"
	oauthRepository "org/api-core/internal/oauth/repository"
	oauthServicePkg "org/api-core/internal/oauth/service"

	billingHandler "org/api-core/internal/billing"
	"org/api-core/internal/queue"
	usageHandlerPkg "org/api-core/internal/usage/handler"
	usageRepository "org/api-core/internal/usage/repository"
	usageServicePkg "org/api-core/internal/usage/service"
	workflowHandlerPkg "org/api-core/internal/workflow/handler"
	workflowRepository "org/api-core/internal/workflow/repository"
	workflowServicePkg "org/api-core/internal/workflow/service"
)

func RegisterRoutes(r *gin.Engine) {
	r.Use(middleware.CORSMiddleware())

	authRoutes(r)
	oauthRoutes(r)
	workflowRoutes(r)
	billingRoutes(r)

	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "server running",
		})
	})
}

func authRoutes(r *gin.Engine) {
	userRepo := authRepository.NewSQLCUserRepository(db.QueriesInstance)
	authService := authServicePkg.NewAuthService(userRepo)
	authHandler := authHandlerPkg.NewAuthHandler(authService)
	googleAuthHandler := authHandlerPkg.NewGoogleAuthHandler(authService)

	r.POST("/signup", authHandler.Signup)
	r.POST("/login", authHandler.Login)
	r.GET("/me", middleware.AuthMiddleware(), authHandler.Me)
	r.GET("/verify-email", authHandler.VerifyEmail)
	r.POST("/forgot-password", authHandler.ForgotPassword)
	r.POST("/reset-password", authHandler.ResetPassword)
	r.POST("/refresh", authHandler.RefreshToken)
	r.POST("/logout", authHandler.Logout)

	r.GET("/auth/google/start", googleAuthHandler.GoogleAuthStart)
	r.GET("/auth/google/callback", googleAuthHandler.GoogleAuthCallback)

	r.GET("/test-billing", handler.TestBilling)
}

func oauthRoutes(r *gin.Engine) {
	oauthRepo := oauthRepository.NewSQLCConnectedAccountRepository(db.QueriesInstance)
	oauthService := oauthServicePkg.NewOAuthService(oauthRepo)
	oauthHandler := oauthHandlerPkg.NewOAuthHandler(oauthService)

	r.GET("/oauth/google/start", middleware.AuthMiddleware(), oauthHandler.GoogleStart)
	r.GET("/oauth/google/callback", oauthHandler.GoogleCallback)
}

func workflowRoutes(r *gin.Engine) {
	workflowRepo := workflowRepository.NewSQLCWorkflowRepository(db.QueriesInstance)

	workflowQueue := queue.NewRedisQueue(tokenstore.RDB)

	workflowService := workflowServicePkg.NewWorkflowService(
		workflowRepo,
		workflowQueue,
	)

	workflowWorker := worker.NewWorkflowWorker(
		workflowQueue,
		workflowService,
	)

	go workflowWorker.Start(context.Background())

	usageRepo := usageRepository.NewPostgresRepository(db.QueriesInstance)
	usageService := usageServicePkg.NewService(usageRepo)

	workflowHandler := workflowHandlerPkg.NewWorkflowHandler(
		workflowService,
		usageService,
	)

	r.POST("/webhooks/:webhookID", workflowHandler.HandleWebhookTrigger)
	r.GET("/workflows/:id/events", workflowHandler.StreamWorkflowEvents)

	auth := r.Group("/", middleware.AuthMiddleware())

	auth.POST("/workflows", workflowHandler.CreateWorkflow)
	auth.GET("/workflows", workflowHandler.ListWorkflows)
	auth.DELETE("/workflows/:id", workflowHandler.DeleteWorkflow)

	auth.POST("/workflows/:id/steps", workflowHandler.CreateStep)

	auth.PUT("/workflows/:id/steps", workflowHandler.SaveWorkflowSteps)
	auth.GET("/workflows/:id/steps", workflowHandler.GetWorkflowSteps)

	auth.PUT("/workflows/:id/schedule", workflowHandler.UpdateWorkflowSchedule)
	auth.GET("/workflows/:id/schedule", workflowHandler.GetWorkflowSchedule)

	auth.POST("/workflows/:id/run", workflowHandler.RunWorkflow)
	auth.GET("/workflows/:id/runs", workflowHandler.ListWorkflowRuns)
	auth.GET("/workflow-runs/:id/steps", workflowHandler.ListWorkflowStepRuns)

	auth.GET("/workflows/:id/edges", workflowHandler.GetWorkflowEdges)
}

func billingRoutes(r *gin.Engine) {
	usageRepo := usageRepository.NewPostgresRepository(db.QueriesInstance)
	usageService := usageServicePkg.NewService(usageRepo)
	usageHandler := usageHandlerPkg.NewUsageHandler(usageService)

	billing := r.Group("/billing")
	billing.Use(middleware.AuthMiddleware())

	billing.POST("/checkout", billingHandler.CreateCheckoutSession)
	billing.GET("/subscription", billingHandler.GetSubscription)
	billing.GET("/usage", usageHandler.GetUsage)
	billing.POST("/portal", billingHandler.CreatePortalSession)
}
