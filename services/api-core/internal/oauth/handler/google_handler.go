package handler

import (
	"net/http"

	"org/api-core/internal/oauth/provider"
	"org/api-core/internal/oauth/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/oauth2"
)

type OAuthHandler struct {
	oauthService *service.OAuthService
}

func NewOAuthHandler(oauthService *service.OAuthService) *OAuthHandler {
	return &OAuthHandler{
		oauthService: oauthService,
	}
}


func (h *OAuthHandler) GoogleStart(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "unauthorized"})
		return
	}

	state := userID.(string)

	config := provider.GoogleAuthConfig()

	url := config.AuthCodeURL(
		state,
		oauth2.AccessTypeOffline,
		oauth2.ApprovalForce,
	)

	c.Redirect(http.StatusTemporaryRedirect, url)
}


func (h *OAuthHandler) GoogleCallback(c *gin.Context) {
	code := c.Query("code")
	state := c.Query("state")

	if code == "" || state == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "code or state missing"})
		return
	}

	userID, err := uuid.Parse(state)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid state"})
		return
	}

	config := provider.GoogleAuthConfig()

	token, err := config.Exchange(c.Request.Context(), code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "failed to exchange token",
			"error":   err.Error(),
		})
		return
	}

	err = h.oauthService.SaveGoogleAccount(
		c.Request.Context(),
		userID,
		token,
		config.Scopes,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "failed to save google account",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "google account connected successfully",
	})
}