package application

import (
	"context"
	"time"

	"github.com/yourusername/userapi/internal/domain"
	"github.com/yourusername/userapi/internal/ports/repository"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
)

// UserService handles business logic for user operations
type UserService struct {
	userRepo repository.UserRepository
}

// NewUserService creates a new user service
func NewUserService(userRepo repository.UserRepository) *UserService {
	return &UserService{userRepo: userRepo}
}

// CreateUser creates a new user with hashed password
func (s *UserService) CreateUser(ctx context.Context, name, email, password string) (*domain.User, error) {
	// Check if user with email already exists
	existingUser, err := s.userRepo.FindByEmail(ctx, email)
	if err == nil && existingUser != nil {
		return nil, domain.ErrEmailAlreadyExists
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// Create new user
	user := domain.NewUser(name, email, string(hashedPassword))
	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

// GetUserByID retrieves a user by ID
func (s *UserService) GetUserByID(ctx context.Context, id string) (*domain.User, error) {
	return s.userRepo.FindByID(ctx, id)
}

// GetAllUsers retrieves all users
func (s *UserService) GetAllUsers(ctx context.Context) ([]*domain.User, error) {
	return s.userRepo.FindAll(ctx)
}

// UpdateUser updates a user's information
func (s *UserService) UpdateUser(ctx context.Context, id, name, email string) (*domain.User, error) {
	// Check if user exists
	user, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Check if email is being updated and is unique
	if email != user.Email {
		existingUser, err := s.userRepo.FindByEmail(ctx, email)
		if err == nil && existingUser != nil {
			return nil, domain.ErrEmailAlreadyExists
		}
	}

	// Update user
	user.Name = name
	user.Email = email

	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

// DeleteUser deletes a user
func (s *UserService) DeleteUser(ctx context.Context, id string) error {
	return s.userRepo.Delete(ctx, id)
}

// CountUsers returns the total number of users
func (s *UserService) CountUsers(ctx context.Context) (int64, error) {
	return s.userRepo.Count(ctx)
}
