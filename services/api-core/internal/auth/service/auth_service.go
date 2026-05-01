package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"org/api-core/internal/auth/tokenStore"
	"org/api-core/internal/auth/repository"
	"org/api-core/internal/auth/security"
	"org/api-core/internal/db"
	"strings"
	"time"
    authEmail "org/api-core/internal/auth/email"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	userRepo repository.UserRepository
}

func NewAuthService(userRepo repository.UserRepository) *AuthService {
	return &AuthService{
		userRepo: userRepo,
	}
}

type TokenPair struct {
	AccessToken  string
	RefreshToken string
}

func (s *AuthService) Signup(ctx context.Context, email, password string) (string, error) {
	existing, _ := s.userRepo.GetUserByEmail(ctx, email)
	if existing.Email != "" {
		return "", errors.New("email already registered")
	}

	if len(password) < 6 {
		return "", errors.New("password must be at least 6 characters")
	}

	hashedPassword, err := security.HashPassword(password)
	if err != nil {
		return "", err
	}

	userID := uuid.New()

	_, err = s.userRepo.CreateUser(ctx, db.CreateUserParams{
		ID:         userID,
		Email:      email,
		Password:   hashedPassword,
		IsVerified: false,
	})
	if err != nil {
		return "", err
	}

	token, err := security.GenerateAccessToken(userID.String())
	if err != nil {
		return "", err
	}

	err = tokenstore.StoreVerificationToken(token, userID.String(), 24*time.Hour)
	if err != nil {
		return "", err
	}

	err = authEmail.SendVerificationEmail(email, token)
	if err != nil {
		return "", err
	}

	return token, nil
}

func (s *AuthService) Login(ctx context.Context, email, password string) (*TokenPair, error) {

	// 🔍 Get user
	user, err := s.userRepo.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	// 🔐 Check password
	err = security.CheckPassword(user.Password, password)
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	// 🚫 Check email verification
	if !user.IsVerified {
		return nil, errors.New("please verify your email first")
	}

	accessToken, err := security.GenerateAccessToken(user.ID.String())
	if err != nil {
		return nil, err
	}

	refreshToken, err := security.GenerateRefreshToken(user.ID.String())
	if err != nil {
		return nil, err
	}

	err = tokenstore.StoreRefreshToken(ctx, refreshToken, user.ID.String())
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *AuthService) VerifyEmail(ctx context.Context, token string) error {
	userID, err := tokenstore.GetUserIDByToken(token)
	if err != nil {
		return errors.New("invalid or expired token")
	}

	parsedID, err := uuid.Parse(userID)
	if err != nil {
		return err
	}

	err = s.userRepo.VerifyUser(ctx, parsedID)
	if err != nil {
		return err
	}

	err = tokenstore.DeleteToken(token)
	if err != nil {
		return err
	}
	return nil
}

func (s *AuthService) ForgotPassword(c context.Context, email string) error {
	user, err := s.userRepo.GetUserByEmail(c, email)
	if err != nil {
		return nil
	}

	token, err := security.GenerateAccessToken(user.ID.String())
	if err != nil {
		return err
	}

	err = tokenstore.StorePasswordResetToken(
		c,
		token,
		user.ID.String(),
		15*time.Minute,
	)
	if err != nil {
		return err
	}

	err = authEmail.SendPasswordResetEmail(user.Email, token)
	if err != nil {
		return err
	}
	return nil
}

func (s *AuthService) ResetPassword(c context.Context, token string, newPassword string) error {
	fmt.Println("RESET TOKEN:", token)
	token = strings.TrimSpace(token)

	userId, err := tokenstore.GetPasswordResetToken(c, token)
	if err != nil {
		fmt.Println("REDIS ERROR:", err)
		return errors.New("invalid or expired token")
	}

	fmt.Println("USER ID FROM REDIS:", userId)

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		fmt.Println("HASH ERROR:", err)
		return err
	}

	userUUID, err := uuid.Parse(userId)
	if err != nil {
		fmt.Println("UUID ERROR:", err)
		return err
	}

	err = s.userRepo.UpdateUserPassword(c, db.UpdateUserPasswordParams{
		Password: string(hashedPassword),
		ID:       userUUID,
	})
	if err != nil {
		fmt.Println("DB UPDATE ERROR:", err)
		return err
	}

	_ = tokenstore.DeletePasswordResetToken(c, token)

	return nil
}

func (s *AuthService) RefreshAccessToken(ctx context.Context, refreshToken string) (*TokenPair, error) {
	refreshToken = strings.TrimSpace(refreshToken)

	// 1. check if token exists
	userID, err := tokenstore.GetRefreshToken(ctx, refreshToken)
	if err != nil {
		return nil, errors.New("invalid refresh token")
	}

	// 2. DELETE old refresh token (rotation)
	_ = tokenstore.DeleteRefreshToken(ctx, refreshToken)

	// 3. generate new tokens
	newAccessToken, err := security.GenerateAccessToken(userID)
	if err != nil {
		return nil, err
	}

	newRefreshToken, err := security.GenerateRefreshToken(userID)
	if err != nil {
		return nil, err
	}

	// 4. store new refresh token
	err = tokenstore.StoreRefreshToken(ctx, newRefreshToken, userID)
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
	}, nil
}

func (s *AuthService) Logout(ctx context.Context, refreshToken string) error {
	refreshToken = strings.TrimSpace(refreshToken)

	return tokenstore.DeleteRefreshToken(ctx, refreshToken)
}



func (s *AuthService) GetOrCreateGoogleUser(ctx context.Context, email string) (*TokenPair, error) {

	user, err := s.userRepo.GetUserByEmail(ctx, email)

	if err != nil {
		if err == sql.ErrNoRows {
			userID := uuid.New()

			user, err = s.userRepo.CreateUser(ctx, db.CreateUserParams{
				ID:         userID,
				Email:      email,
				Password:   "",
				IsVerified: true,
			})
			if err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	accessToken, err := security.GenerateAccessToken(user.ID.String())
	if err != nil {
		return nil, err
	}

	refreshToken, err := security.GenerateRefreshToken(user.ID.String())
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}