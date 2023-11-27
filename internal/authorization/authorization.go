package authorization

import (
	"fmt"
	"graduation/internal/logger"
	"net/http"
	"strconv"

	"github.com/golang-jwt/jwt/v4"
)

type Claims struct {
	jwt.RegisteredClaims
	UserID int
}

func getUserID(secretKey, tokenString string) (int, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims,
		func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return []byte(secretKey), nil
		})
	if err != nil {
		return 0, fmt.Errorf("cannot pars: %v", err)
	}

	if !token.Valid {
		return 0, fmt.Errorf("token is not valid: %v", err)
	}

	return claims.UserID, nil
}

func AuthorizationMiddleware(secretKey string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie("Authorization")
			if err != nil {
				logger.Error("cookies do not contain a token: %v", err)
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			id, err := getUserID(secretKey, cookie.Value)
			if err != nil {
				logger.Error("token does not pass validation")
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			r.Header.Set("User_id", strconv.Itoa(id))

			next.ServeHTTP(w, r)
		})
	}
}
