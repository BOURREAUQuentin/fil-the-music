package server

import (
	"context"
	pb "github.com/BOURREAUQuentin/game/api/proto/v1"
	"github.com/BOURREAUQuentin/game/internal/core/ports"
)

type GameServer struct {
	pb.UnimplementedGameServiceServer
	QuizRepo    ports.QuizRepository
	SessionRepo ports.SessionRepository
	UserRepo    ports.UserRepository
	CatalogRepo ports.CatalogRepository
}

func NewGameServer(quizRepo ports.QuizRepository, sessionRepo ports.SessionRepository, userRepo ports.UserRepository, catalogRepo ports.CatalogRepository) *GameServer {
	return &GameServer{
		QuizRepo:    quizRepo,
		SessionRepo: sessionRepo,
		UserRepo:    userRepo,
		CatalogRepo: catalogRepo,
	}
}

func (s *GameServer) cleanupOldSession(ctx context.Context, userID string) {
	existingSession, _ := s.SessionRepo.GetSessionByUserId(ctx, userID)
	if existingSession != nil {
		// Check associated quiz
		oldQuiz, err := s.QuizRepo.FindQuizByID(ctx, existingSession.QuizId)
		if err == nil && oldQuiz != nil {
			if oldQuiz.Type == "random" || oldQuiz.Type == "genre" || oldQuiz.Type == "for_you" || oldQuiz.Type == "for_you_fallback" {
				_ = s.QuizRepo.DeleteQuiz(ctx, oldQuiz.ID)
			}
		}
		s.SessionRepo.DeleteSession(ctx, existingSession)
	}
}
