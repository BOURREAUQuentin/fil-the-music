package server

import (
	pb "github.com/BOURREAUQuentin/game/api/proto/v1"
	"github.com/BOURREAUQuentin/game/internal/core/ports"
)

type GameServer struct {
	pb.UnimplementedGameServiceServer
	QuizRepo    ports.QuizRepository
	SessionRepo ports.SessionRepository
}

func NewGameServer(quizRepo ports.QuizRepository, sessionRepo ports.SessionRepository) *GameServer {
	return &GameServer{
		QuizRepo:    quizRepo,
		SessionRepo: sessionRepo,
	}
}
