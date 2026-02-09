package auth

import (
	"agn-service/internal/logger"
	"context"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
)

type User struct {
	Username string
	Role     string
	Account  string
}

type Claims struct {
	Username string `json:"username"`
	Role     string `json:"role"`
	Account  string `json:"account"`
	IsActive bool   `json:"isActive"`
	jwt.RegisteredClaims
}

func MiddlewareJWT(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		start := time.Now()

		secret := os.Getenv("SECRET_KEY_API")
		if secret == "" {
			http.Error(w, "SECRET_KEY_API not set", http.StatusInternalServerError)
			logger.Log.Error("jwt_secret_missing")
			return
		}

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Missing Authorization header", http.StatusUnauthorized)
			logger.Log.Warn("jwt_missing_header",
				zap.String("path", r.URL.Path),
			)
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "Invalid Authorization format", http.StatusUnauthorized)
			logger.Log.Warn("jwt_invalid_format",
				zap.String("auth", authHeader),
			)
			return
		}

		tokenStr := parts[1]

		token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (interface{}, error) {
			return []byte(secret), nil
		})

		if err != nil || !token.Valid {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			logger.Log.Warn("jwt_invalid",
				zap.Error(err),
			)
			return
		}

		claims := token.Claims.(*Claims)

		if !claims.IsActive {
			http.Error(w, "User is not active", http.StatusForbidden)
			logger.Log.Warn("jwt_user_inactive",
				zap.String("user", claims.Username),
			)
			return
		}

		user := &User{
			Username: claims.Username,
			Role:     claims.Role,
			Account:  claims.Account,
		}

		// crear contexto ANTES de llamar al handler
		ctx := context.WithValue(r.Context(), "user", user)

		// llamar UNA sola vez
		next.ServeHTTP(w, r.WithContext(ctx))

		// 👉 log después de atender la request
		logger.Log.Info("http_request",
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
			zap.String("remote", r.RemoteAddr),
			zap.Duration("elapsed_ms", time.Since(start)),
			zap.String("user", user.Username),
		)
	})
}

func FromContext(ctx context.Context) *User {
	user, ok := ctx.Value("user").(*User)
	if !ok {
		return nil
	}
	return user
}
