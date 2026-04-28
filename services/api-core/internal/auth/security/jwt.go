package security

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var secret = []byte("supersecret")

func GenerateToken(userId string) (string, error) {
	claims := jwt.MapClaims{
        "user_id":userId,
        "exp":time.Now().Add(time.Minute * 15).Unix(),
    }
    
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString(secret)
}