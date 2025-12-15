package repository

import (
	"context"

	"github.com/BOURREAUQuentin/game/internal/core/domain"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type UserRepository struct{}

func NewUserRepository() *UserRepository {
	return &UserRepository{}
}

func (u UserRepository) GetUserById(ctx context.Context, id string) (*domain.User, error) {
	if id != "1" {
		return nil, status.Error(codes.PermissionDenied, "User not found")
	}

	return &domain.User{
		ID:       "1",
		Username: "username",
		Role:     domain.RoleAdmin,
	}, nil
}
