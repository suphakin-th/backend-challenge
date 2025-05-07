package http

import (
	"context"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/yourusername/userapi/pkg/auth"
)

// Middleware type
type Middleware func(http.Handler) http.Handler

// LoggingMiddleware logs HTTP method, path, and execution time
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		
		// Call the next handler
		next.ServeHTTP(w, r)
		
		// Log the request details
		log.Printf(
			"Method: %s\tPath: %s\tTime: %v",
			r.Method,
			r.URL.Path,
			time.Since(start),
		)
	})
}

// AuthMiddleware validates JWT tokens
func AuthMiddleware(jwtAuth *auth.JWTAuth) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get token from Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				respondWithError(w, http.StatusUnauthorized, "Authorization header is required")
				return
			}
			
			// Check if the header format is valid
			bearerToken := strings.Split(authHeader, " ")
			if len(bearerToken) != 2 || strings.ToLower(bearerToken[0]) != "bearer" {
				respondWithError(w, http.StatusUnauthorized, "Invalid authorization format. Format: Bearer {token}")
				return
			}
			
			// Validate token
			claims, err := jwtAuth.ValidateToken(bearerToken[1])
			if err != nil {
				respondWithError(w, http.StatusUnauthorized, "Invalid or expired token")
				return
			}
			
			// Set user ID in context
			ctx := context.WithValue(r.Context(), "userID", claims.UserID)
			ctx = context.WithValue(ctx, "email", claims.Email)
			
			// Call the next handler with our new context
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}