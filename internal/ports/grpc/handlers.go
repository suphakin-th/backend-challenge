package grpc

import (
	"context"

	"github.com/yourusername/userapi/internal/application"
	"github.com/yourusername/userapi/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// UserServiceHandler implements the gRPC service
type UserServiceHandler struct {
	proto.UnimplementedUserServiceServer
	userService *application.UserService
	authService *application.AuthService
}

// NewUserServiceHandler creates a new gRPC handler
func NewUserServiceHandler(userService *application.UserService, authService *application.AuthService) *UserServiceHandler {
	return &UserServiceHandler{
		userService: userService,
		authService: authService,
	}
}

// CreateUser implements the gRPC CreateUser method
func (h *UserServiceHandler) CreateUser(ctx context.Context, req *proto.CreateUserRequest) (*proto.UserResponse, error) {
	user, err := h.authService.Register(ctx, req.Name, req.Email, req.Password)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create user: %v", err)
	}

	return &proto.UserResponse{
		Id:        user.ID.Hex(),
		Name:      user.Name,
		Email:     user.Email,
		CreatedAt: timestamppb.New(user.CreatedAt).String(),
	}, nil
}

// GetUser implements the gRPC GetUser method
func (h *UserServiceHandler) GetUser(ctx context.Context, req *proto.GetUserRequest) (*proto.UserResponse, error) {
	user, err := h.userService.GetUserByID(ctx, req.Id)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "user not found: %v", err)
	}

	return &proto.UserResponse{
		Id:        user.ID.Hex(),
		Name:      user.Name,
		Email:     user.Email,
		CreatedAt: timestamppb.New(user.CreatedAt).String(),
	}, nil
}
