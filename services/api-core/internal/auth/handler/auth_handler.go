package handler

import (
	"Org/utils/response"
	"fmt"
	"org/api-core/internal/auth/service"

	"github.com/gin-gonic/gin"
)

func Signup(c *gin.Context) {
	var body struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "invalid input", err.Error())
		return
	}

	if body.Email == "" || body.Password == "" {
		response.BadRequest(c, "email and password required", nil)
		return
	}

	authService := service.NewAuthService()

	token, err := authService.Signup(c, body.Email, body.Password)
	if err != nil {
		response.BadRequest(c, err.Error(), nil)
		return
	}

	err = service.SendVerificationEmail(body.Email, token)
	if err != nil {
		response.InternalServerError(c, "failed to send verification email", err.Error())
		return
	}

	response.Created(c, "user created. please check your email to verify account", nil)
}

func Login(c *gin.Context) {
	var body struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "invalid input", err.Error())
		return
	}

	if body.Email == "" || body.Password == "" {
		response.BadRequest(c, "email and password required", nil)
		return
	}

	authService := service.NewAuthService()

	token, err := authService.Login(c, body.Email, body.Password)
	if err != nil {
		response.Unauthorized(c, err.Error())
		return
	}

	response.OK(c, "login successful", gin.H{
		"token": token,
	})
}

func Me(c *gin.Context) {
	userID, _ := c.Get("user_id")

	response.OK(c, "you are authenticated", gin.H{
		"user_id": userID,
	})
}

func VerifyEmail(c *gin.Context) {
	token := c.Query("token")

	if token == "" {
		response.BadRequest(c, "token is required", nil)
		return
	}

	authService := service.NewAuthService()

	err := authService.VerifyEmail(c, token)
	if err != nil {
		response.BadRequest(c, c.Err().Error(), nil)
		return
	}

	response.OK(c, "email verified successfully", nil)
}

func ForgotPassword(c *gin.Context) {
	var body struct {
		Email string `json:"email"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "invalid input", err.Error())
		return
	}

	authService := service.NewAuthService()

	err := authService.ForgotPassword(c, body.Email)
	if err != nil {
		response.OK(c, "if this email exists, reset link has been sent", nil)
		return
	}

	response.OK(c, "if this email exists, reset link has been sent", nil)
}

func ResetPassword(c *gin.Context) {
	var body struct {
		Token       string `json:"token"`
		NewPassword string `json:"new_password"`
	}
    
    
	if err := c.ShouldBindJSON(&body); err != nil {
        response.BadRequest(c, "invalid input", err.Error())
		return
	}
    fmt.Println("RAW TOKEN:", body.Token)
    fmt.Printf("TOKEN LEN: %d\n", len(body.Token))
    
	if body.Token == "" || body.NewPassword == "" {
		response.BadRequest(c, "token and new password required", nil)
		return
	}

	authService := service.NewAuthService()

	err := authService.ResetPassword(c, body.Token, body.NewPassword)
	if err != nil {
		c.JSON(500, gin.H{
			"status":  500,
			"message": "reset failed",
			"error":   err.Error(),
		})
		return
	}

	response.OK(c, "password reset successfully", nil)
}
