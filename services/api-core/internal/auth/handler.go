package auth

import (
	"Org/utils/response"
	"context"
	"org/api-core/internal/auth/security"
	"org/api-core/internal/db"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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

	hashedPassword, err := security.HashPassword(body.Password)
	if err != nil {
		response.InternalServerError(c, "failed to hash password", err.Error())
		return
	}

	ctx := context.Background()

	user, err := db.QueriesInstance.CreateUser(ctx, db.CreateUserParams{
		ID:       uuid.New(),
		Email:    body.Email,
		Password: hashedPassword,
	})

	if err != nil {
		response.InternalServerError(c, "failed to create user", err.Error())
		return
	}

	response.Created(c, "user created successfully", gin.H{
		"user_id": user.ID,
		"email":   user.Email,
	})
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

	ctx := context.Background()

	user, err := db.QueriesInstance.GetUserByEmail(ctx, body.Email)
	if err != nil {
		response.Unauthorized(c, "invalid credentials")
		return
	}

	err = security.CheckPassword(user.Password, body.Password)
	if err != nil {
		response.Unauthorized(c, "invalid credentials")
	}

	token, err := security.GenerateToken(user.ID.String())
	if err != nil {
		response.InternalServerError(c, "failed to generate token", err.Error())
		return
	}

	response.OK(c, "login successfull", gin.H{
		"token": token,
	})
}

func Me(c *gin.Context) {
	userID, _ := c.Get("user_id")

	response.OK(c, "you are authenticated", gin.H{
		"user_id": userID,
	})
}
