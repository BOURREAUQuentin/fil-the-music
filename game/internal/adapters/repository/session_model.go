package repository

import (
	"time"

	"github.com/BOURREAUQuentin/game/internal/core/domain"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type SessionMongoDoc struct {
	ID           primitive.ObjectID `bson:"_id,omitempty"`
	UserId       string             `bson:"user_id"`
	QuizId       string             `bson:"quiz_id"`
	JoinedAt     time.Time          `bson:"joined_at"`
	LastActivity time.Time          `bson:"last_activity"`
}

func (s *SessionMongoDoc) ToDomain() domain.Session {
	return domain.Session{
		ID:           s.ID.Hex(),
		UserId:       s.UserId,
		QuizId:       s.QuizId,
		JoinedAt:     s.JoinedAt,
		LastActivity: s.LastActivity,
	}
}
