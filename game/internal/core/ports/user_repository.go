package ports

import (
	"context"

	"github.com/BOURREAUQuentin/game/internal/core/domain"
)

type UserRepository interface {
	GetUserById(ctx context.Context, id string) (*domain.User, error)
}
