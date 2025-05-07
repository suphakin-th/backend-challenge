package application

import (
	"context"
	"time"

	"github.com/yourusername/userapi/internal/domain"
	"github.com/yourusername/userapi/internal/ports/repository"
	"github.com/yourusername/userapi/pkg/auth"
	"golang.org/x/crypto/bcrypt"
)

// AuthService handles authentication logic
type AuthService struct {
	userRepo repository.UserRepository
	jwtAuth  *auth.JWTAuth
}

// NewAuthService creates a new authentication service
func NewAuthService(userRepo repository.UserRepository, jwtAuth *auth.JWTAuth) *AuthService {
	return &AuthService{
		userRepo: userRepo,
		jwtAuth:  jwtAuth,
	}
}

// Register registers a new user
func (s *AuthService) Register(ctx context.Context, name, email, password string) (*domain.User, error) {
	// Use UserService to create user
	userService := NewUserService(s.userRepo)
	return userService.CreateUser(ctx, name, email, password)
}

// Login authenticates a user and returns a JWT token
func (s *AuthService) Login(ctx context.Context, email, password string) (string, error) {
	// Find user by email
	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return "", domain.ErrInvalidCredentials
	}

	// Compare password
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return "", domain.ErrInvalidCredentials
	}

	// Generate JWT token
	token, err := s.jwtAuth.GenerateToken(user.ID.Hex(), user.Email)
	if err != nil {
		return "", err
	}

	return token, nil
}
