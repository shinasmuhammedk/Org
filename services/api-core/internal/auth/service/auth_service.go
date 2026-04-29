package service

import (
	"context"
	"errors"
	"fmt"
	"org/api-core/internal/auth/cache"
	"org/api-core/internal/auth/security"
	"org/api-core/internal/db"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct{}

func NewAuthService() *AuthService {
	return &AuthService{}
}

func (s *AuthService) Signup(ctx context.Context, email, password string) (string, error) {

	// 🔴 Check existing user
	existing, _ := db.QueriesInstance.GetUserByEmail(ctx, email)
	if existing.Email != "" {
		return "", errors.New("email already registered")
	}

	// 🔐 Validate password (basic)
	if len(password) < 6 {
		return "", errors.New("password must be at least 6 characters")
	}

	// 🔐 Hash password
	hashedPassword, err := security.HashPassword(password)
	if err != nil {
		return "", err
	}

	// 🆔 Create user
	userID := uuid.New()

	_, err = db.QueriesInstance.CreateUser(ctx, db.CreateUserParams{
		ID:         userID,
		Email:      email,
		Password:   hashedPassword,
		IsVerified: false,
	})
	if err != nil {
		return "", err
	}

	// 🔑 Generate verification token
	token, err := security.GenerateToken(userID.String())
	if err != nil {
		return "", err
	}
	// 🧠 Store in Redis (TTL 15 min)
	err = cache.StoreVerificationToken(token, userID.String(), 15*time.Minute)
	if err != nil {
		return "", err
	}

	// 📧 Send email (for now just log)
	verifyLink := "http://localhost:8080/verify-email?token=" + token
	println("VERIFY LINK:", verifyLink)

	return token, nil
}

func (s *AuthService) Login(ctx context.Context, email, password string) (string, error) {

	// 🔍 Get user
	user, err := db.QueriesInstance.GetUserByEmail(ctx, email)
	if err != nil {
		return "", errors.New("invalid credentials")
	}

	// 🔐 Check password
	err = security.CheckPassword(user.Password, password)
	if err != nil {
		return "", errors.New("invalid credentials")
	}

	// 🚫 Check email verification
	if !user.IsVerified {
		return "", errors.New("please verify your email first")
	}

	// 🔑 Generate JWT
	token, err := security.GenerateToken(user.ID.String())
	if err != nil {
		return "", err
	}

	return token, nil
}

func (s *AuthService) VerifyEmail(ctx context.Context, token string) error {
	userID, err := cache.GetUserIDByToken(token)
	if err != nil {
		return errors.New("invalid or expired token")
	}

	parsedID, err := uuid.Parse(userID)
	if err != nil {
		return err
	}

	err = db.QueriesInstance.VerifyUser(ctx, parsedID)
	if err != nil {
		return err
	}

	err = cache.DeleteToken(token)
	if err != nil {
		return err
	}
	return nil
}

func (s *AuthService) ForgotPassword(c context.Context, email string) error {
	user, err := db.QueriesInstance.GetUserByEmail(c, email)
	if err != nil {
		return nil
	}

	token, err := security.GenerateToken(user.ID.String())
	if err != nil {
		return err
	}

	err = cache.StorePasswordResetToken(
		c,
		token,
		user.ID.String(),
		15*time.Minute,
	)
	if err != nil {
		return err
	}

	err = SendPasswordResetEmail(user.Email, token)
	if err != nil {
		return err
	}
	return nil
}

func (s *AuthService) ResetPassword(c context.Context, token string, newPassword string) error {
	fmt.Println("RESET TOKEN:", token)
	token = strings.TrimSpace(token)

	userId, err := cache.GetPasswordResetToken(c, token)
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

	err = db.QueriesInstance.UpdateUserPassword(c, db.UpdateUserPasswordParams{
		Password: string(hashedPassword),
		ID:       userUUID,
	})
	if err != nil {
		fmt.Println("DB UPDATE ERROR:", err)
		return err
	}

	_ = cache.DeletePasswordResetToken(c, token)

	return nil
}
