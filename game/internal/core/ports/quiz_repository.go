package ports

import (
	"context"

	"github.com/BOURREAUQuentin/game/internal/core/domain"
)

type QuizRepository interface {
	FindAllQuizzes(ctx context.Context) ([]domain.Quiz, error)
	FindQuizByID(ctx context.Context, id string) (*domain.Quiz, error)
}
