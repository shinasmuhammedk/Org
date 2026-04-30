package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	authHandlerPkg "org/api-core/internal/auth/handler"
	"org/api-core/internal/auth/cache"
	"org/api-core/internal/auth/middleware"
	authRepository "org/api-core/internal/auth/repository"
	authServicePkg "org/api-core/internal/auth/service"

	"org/api-core/internal/db"

	oauthHandlerPkg "org/api-core/internal/oauth/handler"
	oauthRepository "org/api-core/internal/oauth/repository"
	oauthServicePkg "org/api-core/internal/oauth/service"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found")
	}

	db.Init()
	cache.InitRedis("127.0.0.1:6379")

	// 🔐 Auth dependencies
	userRepo := authRepository.NewSQLCUserRepository(db.QueriesInstance)
	authService := authServicePkg.NewAuthService(userRepo)
	authHandler := authHandlerPkg.NewAuthHandler(authService)
	googleAuthHandler := authHandlerPkg.NewGoogleAuthHandler(authService)

	// 🔗 OAuth (connect integrations) dependencies
	oauthRepo := oauthRepository.NewSQLCConnectedAccountRepository(db.QueriesInstance)
	oauthService := oauthServicePkg.NewOAuthService(oauthRepo)
	oauthHandler := oauthHandlerPkg.NewOAuthHandler(oauthService)

	r := gin.Default()
	r.Use(middleware.CORSMiddleware())

	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "server running",
		})
	})

	// 🔐 Auth routes
	r.POST("/signup", authHandler.Signup)
	r.POST("/login", authHandler.Login)
	r.GET("/me", middleware.AuthMiddleware(), authHandler.Me)
	r.GET("/verify-email", authHandler.VerifyEmail)
	r.POST("/forgot-password", authHandler.ForgotPassword)
	r.POST("/reset-password", authHandler.ResetPassword)
	r.POST("/refresh", authHandler.RefreshToken)
	r.POST("/logout", authHandler.Logout)

	// 🔥 Google OAuth SIGNUP / LOGIN
	r.GET("/auth/google/start", googleAuthHandler.GoogleAuthStart)
	r.GET("/auth/google/callback", googleAuthHandler.GoogleAuthCallback)

	// 🔗 Google OAuth CONNECT (for workflows)
	r.GET("/oauth/google/start", middleware.AuthMiddleware(), oauthHandler.GoogleStart)
	r.GET("/oauth/google/callback", oauthHandler.GoogleCallback)

	r.Run(":8080")
}