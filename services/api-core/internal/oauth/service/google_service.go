package service

import (
	"context"
	"database/sql"
	"strings"

	"org/api-core/internal/db"
	"org/api-core/internal/oauth/repository"

	"github.com/google/uuid"
	"golang.org/x/oauth2"
)

type OAuthService struct {
	accountRepo repository.ConnectedAccountRepository
}

func NewOAuthService(accountRepo repository.ConnectedAccountRepository) *OAuthService {
	return &OAuthService{
		accountRepo: accountRepo,
	}
}

func (s *OAuthService) SaveGoogleAccount(
	ctx context.Context,
	userID uuid.UUID,
	token *oauth2.Token,
	scopes []string,
) error {
	_, err := s.accountRepo.UpsertConnectedAccount(ctx, db.UpsertConnectedAccountParams{
		UserID:      userID,
		Provider:    "google",
		AccessToken: token.AccessToken,
		RefreshToken: sql.NullString{
			String: token.RefreshToken,
			Valid:  token.RefreshToken != "",
		},
		ExpiresAt: sql.NullTime{
			Time:  token.Expiry,
			Valid: !token.Expiry.IsZero(),
		},
		Scopes: sql.NullString{
			String: strings.Join(scopes, " "),
			Valid:  len(scopes) > 0,
		},
	})
	return err
}
