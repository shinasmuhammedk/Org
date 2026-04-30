package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"org/api-core/internal/auth/service"
	"org/api-core/internal/oauth/provider"

	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
)

type GoogleAuthHandler struct {
	authService *service.AuthService
}

func NewGoogleAuthHandler(authService *service.AuthService) *GoogleAuthHandler {
	return &GoogleAuthHandler{
		authService: authService,
	}
}

// ✅ START endpoint
func (h *GoogleAuthHandler) GoogleAuthStart(c *gin.Context) {
	config := provider.GoogleAuthConfig()

	url := config.AuthCodeURL(
		"oauth_signup",
		oauth2.AccessTypeOffline,
	)

	c.Redirect(http.StatusTemporaryRedirect, url)
}

// ✅ CALLBACK (this is Step 3 you asked about)
func (h *GoogleAuthHandler) GoogleAuthCallback(c *gin.Context) {
	code := c.Query("code")

	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "code missing"})
		return
	}

	config := provider.GoogleAuthConfig()

	token, err := config.Exchange(context.Background(), code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "token exchange failed"})
		return
	}

	client := config.Client(context.Background(), token)

	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to fetch user"})
		return
	}
	defer resp.Body.Close()

	var userInfo struct {
		ID    string `json:"id"`
		Email string `json:"email"`
		Name  string `json:"name"`
	}

	json.NewDecoder(resp.Body).Decode(&userInfo)

	tokenPair, err := h.authService.GetOrCreateGoogleUser(c.Request.Context(), userInfo.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "auth failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token":  tokenPair.AccessToken,
		"refresh_token": tokenPair.RefreshToken,
	})
}