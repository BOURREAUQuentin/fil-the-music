package server

import (
	"context"

	pb "github.com/BOURREAUQuentin/game/api/proto/v1"
	"github.com/BOURREAUQuentin/game/internal/adapters/repository" // Import ton repo
	"github.com/BOURREAUQuentin/game/internal/core/domain"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type GameServer struct {
	pb.UnimplementedGameServiceServer
	Repo repository.MongoRepository
}

func (s *GameServer) GetQuizzes(ctx context.Context, req *pb.GetQuizzesRequest) (*pb.GetQuizzesResponse, error) {
	quizzesDB, err := s.Repo.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	var pbQuizzes []*pb.Quiz
	for _, q := range quizzesDB {
		pbQuizzes = append(pbQuizzes, &pb.Quiz{
			Id:        q.ID,
			Title:     q.Title,
			Type:      q.Type,
			CreatorId: q.CreatorID,
			Status:    q.Status,
			CreatedAt: timestamppb.New(q.CreatedAt),
			Questions: mapQuestionsToPb(q.Questions),
		})
	}

	return &pb.GetQuizzesResponse{Quizzes: pbQuizzes}, nil
}

func mapQuestionsToPb(qs []domain.Question) []*pb.Question {
	var pbQs []*pb.Question
	for _, q := range qs {
		pbQs = append(pbQs, &pb.Question{
			QuestionId: q.QuestionID,
			Text:       q.Text,
		})
	}
	return pbQs
}
