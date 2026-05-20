package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"org/api-core/internal/auth/service"
	"org/api-core/internal/oauth/provider"
	"org/api-core/internal/utils/response"

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

func (h *GoogleAuthHandler) GoogleAuthStart(c *gin.Context) {
	config := provider.GoogleAuthConfig()

	url := config.AuthCodeURL(
		"oauth_signup",
		oauth2.AccessTypeOffline,
	)

	c.Redirect(http.StatusTemporaryRedirect, url)
}

func (h *GoogleAuthHandler) GoogleAuthCallback(c *gin.Context) {
	code := c.Query("code")

	if code == "" {
		response.BadRequest(c, "code missing", nil)
		return
	}

	config := provider.GoogleAuthConfig()

	token, err := config.Exchange(context.Background(), code)
	if err != nil {
		response.InternalServerError(c, "token exchange failed", err.Error())
		return
	}

	client := config.Client(context.Background(), token)

	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		response.InternalServerError(c, "failed to fetch google user", err.Error())
		return
	}
	defer resp.Body.Close()

	var userInfo struct {
		ID    string `json:"id"`
		Email string `json:"email"`
		Name  string `json:"name"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		response.InternalServerError(c, "failed to decode google user", err.Error())
		return
	}

	if userInfo.Email == "" {
		response.BadRequest(c, "google account email missing", nil)
		return
	}

	tokenPair, err := h.authService.GetOrCreateGoogleUser(
		c.Request.Context(),
		userInfo.Email,
	)

	if err != nil {
		response.InternalServerError(c, "google auth failed", err.Error())
		return
	}
	frontendURL := os.Getenv("APP_URL")

	redirectURL := fmt.Sprintf(
		"%s/oauth/callback?access_token=%s&refresh_token=%s",
		frontendURL,
		tokenPair.AccessToken,
		tokenPair.RefreshToken,
	)

	c.Redirect(http.StatusTemporaryRedirect, redirectURL)
}
