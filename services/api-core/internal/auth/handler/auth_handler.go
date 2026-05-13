package handler

import (
	"Org/utils/response"
	"context"
	"fmt"
	"org/api-core/internal/auth/service"
	"org/api-core/internal/billing"

	"github.com/gin-gonic/gin"
	pb "org/api-core/proto"
)

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

func (h *AuthHandler) Signup(c *gin.Context) {
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

	_, err := h.authService.Signup(c.Request.Context(), body.Email, body.Password)
	if err != nil {
		response.BadRequest(c, err.Error(), nil)
		return
	}

	response.Created(c, "user created. please check your email to verify account", nil)
}

func (h *AuthHandler) Login(c *gin.Context) {
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

	tokenPair, err := h.authService.Login(c.Request.Context(), body.Email, body.Password)
	if err != nil {
		response.Unauthorized(c, err.Error())
		return
	}

	response.OK(c, "login successfull", gin.H{
		"access_token":  tokenPair.AccessToken,
		"refresh_token": tokenPair.RefreshToken,
	})
}

func (h *AuthHandler) Me(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "unauthorized")
		return
	}

	response.OK(c, "you are authenticated", gin.H{
		"user_id": userID,
	})
}

func (h *AuthHandler) VerifyEmail(c *gin.Context) {
	token := c.Query("token")

	if token == "" {
		response.BadRequest(c, "token is required", nil)
		return
	}

	err := h.authService.VerifyEmail(c.Request.Context(), token)
	if err != nil {
		response.BadRequest(c, err.Error(), nil)
		return
	}

	response.OK(c, "email verified successfully", nil)
}

func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	var body struct {
		Email string `json:"email"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "invalid input", err.Error())
		return
	}

	err := h.authService.ForgotPassword(c.Request.Context(), body.Email)
	if err != nil {
		response.OK(c, "if this email exists, reset link has been sent", nil)
		return
	}

	response.OK(c, "if this email exists, reset link has been sent", nil)
}

func (h *AuthHandler) ResetPassword(c *gin.Context) {
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

	err := h.authService.ResetPassword(c.Request.Context(), body.Token, body.NewPassword)
	if err != nil {
		response.InternalServerError(c, "reset failed", err.Error())
		return
	}

	response.OK(c, "password reset successfully", nil)
}

func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var body struct {
		RefreshToken string `json:"refresh_token"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "invalid input", err.Error())
		return
	}

	tokenPair, err := h.authService.RefreshAccessToken(c.Request.Context(), body.RefreshToken)
	if err != nil {
		response.Unauthorized(c, "invalid refresh token")
		return
	}

	response.OK(c, "token refreshed", gin.H{
		"access_token":  tokenPair.AccessToken,
		"refresh_token": tokenPair.RefreshToken,
	})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	var body struct {
		RefreshToken string `json:"refresh_token"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "invalid input", err.Error())
		return
	}

	err := h.authService.Logout(c.Request.Context(), body.RefreshToken)
	if err != nil {
		response.InternalServerError(c, "logout failed", err.Error())
		return
	}

	response.OK(c, "logged out successfully", nil)
}

func TestBilling(c *gin.Context) {

	res, err := billing.Client.GetUserSubscription(
		context.Background(),
		&pb.GetUserSubscriptionRequest{
			UserId: "11111111-1111-1111-1111-111111111111",
		},
	)

	if err != nil {
		c.JSON(500, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{
		"plan":   res.Plan,
		"status": res.Status,
	})
}
