package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

const (
	AuthorizationHeader = "Authorization"
	BearerPrefix       = "Bearer "
	UserIDKey          = "user_id"
	UsernameKey        = "username"
)

// JWT auth middleware
func JWTAuth(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip auth for login/register
			if strings.Contains(r.URL.Path, "/auth/login") || 
			   strings.Contains(r.URL.Path, "/auth/register") {
				next.ServeHTTP(w, r)
				return
			}

			authHeader := r.Header.Get(AuthorizationHeader)
			if authHeader == "" {
				http.Error(w, "Missing authorization header", http.StatusUnauthorized)
				return
			}

			if !strings.HasPrefix(authHeader, BearerPrefix) {
				http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
				return
			}

			tokenString := strings.TrimPrefix(authHeader, BearerPrefix)

			// Parse and validate token
			token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
				return []byte(secret), nil
			})

			if err != nil || !token.Valid {
				http.Error(w, "Invalid token", http.StatusUnauthorized)
				return
			}

			// Extract claims
			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				http.Error(w, "Invalid token claims", http.StatusUnauthorized)
				return
			}

			// Add user info to context
			ctx := r.Context()
			if userID, ok := claims["user_id"].(float64); ok {
				ctx = context.WithValue(ctx, UserIDKey, uint(userID))
			}
			if username, ok := claims["username"].(string); ok {
				ctx = context.WithValue(ctx, UsernameKey, username)
			}

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// CORS middleware
func CORS(origin string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			w.Header().Set("Access-Control-Allow-Credentials", "true")

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
