package ports

import (
	"context"

	"github.com/BOURREAUQuentin/game/internal/core/domain"
)

type QuizRepository interface {
	List(ctx context.Context) ([]domain.Quiz, error)
}
