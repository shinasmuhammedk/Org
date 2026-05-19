package app

import (
	"context"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	tokenstore "org/api-core/internal/auth/tokenstore"
	"org/api-core/internal/db"
	"org/api-core/internal/queue"
	"org/api-core/internal/router"
	"org/api-core/internal/scheduler"
	workflowRepository "org/api-core/internal/workflow/repository"
	workflowServicePkg "org/api-core/internal/workflow/service"
)

func New() *gin.Engine {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	db.Init()
	tokenstore.InitRedis("127.0.0.1:6379")

	r := gin.Default()

	router.RegisterRoutes(r)

	workflowRepo := workflowRepository.NewSQLCWorkflowRepository(db.QueriesInstance)
	workflowQueue := queue.NewRedisQueue(tokenstore.RDB)
	workflowService := workflowServicePkg.NewWorkflowService(
		workflowRepo,
		workflowQueue,
	)

	schedulerWorker := scheduler.NewScheduler(workflowService)
	schedulerWorker.Start(context.Background())

	return r
}
