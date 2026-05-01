package router

import (
	"github.com/gin-gonic/gin"

	authHandlerPkg "org/api-core/internal/auth/handler"
	"org/api-core/internal/auth/middleware"
	authRepository "org/api-core/internal/auth/repository"
	authServicePkg "org/api-core/internal/auth/service"

	"org/api-core/internal/db"

	oauthHandlerPkg "org/api-core/internal/oauth/handler"
	oauthRepository "org/api-core/internal/oauth/repository"
	oauthServicePkg "org/api-core/internal/oauth/service"
)

func RegisterRoutes(r *gin.Engine) {
	r.Use(middleware.CORSMiddleware())

	authRoutes(r)
	oauthRoutes(r)

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
}

func oauthRoutes(r *gin.Engine) {
	oauthRepo := oauthRepository.NewSQLCConnectedAccountRepository(db.QueriesInstance)
	oauthService := oauthServicePkg.NewOAuthService(oauthRepo)
	oauthHandler := oauthHandlerPkg.NewOAuthHandler(oauthService)

	r.GET("/oauth/google/start", middleware.AuthMiddleware(), oauthHandler.GoogleStart)
	r.GET("/oauth/google/callback", oauthHandler.GoogleCallback)
}