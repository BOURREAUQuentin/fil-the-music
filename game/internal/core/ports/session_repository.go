package ports

import (
	"context"
	"time"

	"github.com/BOURREAUQuentin/game/internal/core/domain"
)

type SessionRepository interface {
	GetSessions(ctx context.Context) ([]domain.Session, error)
	GetSessionByUserId(ctx context.Context, userId string) (*domain.Session, error)
	CreateSession(ctx context.Context, session *domain.Session) error
	DeleteSession(ctx context.Context, session *domain.Session) error
	UpdateActivity(ctx context.Context, sessionId string, lastActivity time.Time) error
}
