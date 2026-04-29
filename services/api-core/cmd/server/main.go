package main

import (
	"log"

	"org/api-core/internal/auth/cache"
	"org/api-core/internal/auth/handler"
	"org/api-core/internal/auth/middleware"
	"org/api-core/internal/db"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found")
	}

	db.Init()
	cache.InitRedis("127.0.0.1:6379")

	r := gin.Default()

	r.Use(middleware.CORSMiddleware())

	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "server running",
		})
	})

	r.POST("/signup", handler.Signup)
	r.POST("/login", handler.Login)
	r.GET("/me", middleware.AuthMiddleware(), handler.Me)
	r.GET("/verify-email", handler.VerifyEmail)
    r.POST("/forgot-password", handler.ForgotPassword)
    r.POST("/reset-password", handler.ResetPassword)

	r.Run(":8080")
}