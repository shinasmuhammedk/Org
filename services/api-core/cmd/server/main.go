package main

import (
	"org/api-core/internal/auth"
	"org/api-core/internal/auth/middleware"
	"org/api-core/internal/db"

	"github.com/gin-gonic/gin"
)

func main() {
    
    db.Init()
    
    
	r := gin.Default()
    
    r.Use(middleware.CORSMiddleware())

	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "server running",
		})
	})

	r.POST("/signup", auth.Signup)
    r.POST("/login", auth.Login)
    r.GET("/me", middleware.AuthMiddleware(),auth.Me)

	r.Run(":8080")
}
