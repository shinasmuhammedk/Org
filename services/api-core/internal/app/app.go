package app

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"org/api-core/internal/db"
	"org/api-core/internal/router"
	tokenstore "org/api-core/internal/auth/tokenStore"
)

func New() *gin.Engine {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	db.Init()
	tokenstore.InitRedis("127.0.0.1:6379")

	r := gin.Default()

	router.RegisterRoutes(r)

	return r
}