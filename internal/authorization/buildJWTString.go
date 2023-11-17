package authorization

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

func BuildJWTString(secretKey string, tokenEXP time.Duration, id int) (string, error) {
	// id := utils.GenerateString()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(tokenEXP)),
		},
		UserID: id,
	})

	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", fmt.Errorf("cannot get token: %v", err)
	}

	return tokenString, nil
}
