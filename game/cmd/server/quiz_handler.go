package server

import (
	"context"
	"time"

	pb "github.com/BOURREAUQuentin/game/api/proto/v1"
	"github.com/BOURREAUQuentin/game/internal/core/domain"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *GameServer) GetQuizzes(ctx context.Context, req *pb.GetQuizzesRequest) (*pb.GetQuizzesResponse, error) {
	quizzesDB, err := s.QuizRepo.FindAllQuizzes(ctx)
	if err != nil {
		return nil, err
	}

	// Format datas
	var pbQuizzes []*pb.Quiz
	for _, q := range quizzesDB {
		pbQuizzes = append(pbQuizzes, &pb.Quiz{
			Id:        q.ID,
			Title:     q.Title,
			Type:      q.Type,
			CreatorId: q.CreatorID,
			CreatedAt: timestamppb.New(q.CreatedAt),
			Questions: mapQuestionsToPb(q.Questions),
		})
	}

	return &pb.GetQuizzesResponse{Quizzes: pbQuizzes}, nil
}

func (s *GameServer) CreateQuiz(ctx context.Context, req *pb.CreateQuizRequest) (*pb.CreateQuizResponse, error) {
	// TODO: Add validation if necessary (e.g., check if creator exists)

	now := time.Now()
	var questions []domain.Question
	for _, q := range req.Questions {
		// Generate ID for question if needed or use provided one
		qID := q.QuestionId
		if qID == "" {
			qID = primitive.NewObjectID().Hex()
		}
		questions = append(questions, domain.Question{
			QuestionID: qID,
			Text:       q.Text,
		})
	}

	quiz := &domain.Quiz{
		Title:     req.Title,
		Type:      req.Type,
		CreatorID: req.CreatorId,
		CreatedAt: now,
		Questions: questions,
	}

	err := s.QuizRepo.CreateQuiz(ctx, quiz)
	if err != nil {
		return nil, err
	}

	return &pb.CreateQuizResponse{
		Quiz: &pb.Quiz{
			Id:        quiz.ID,
			Title:     quiz.Title,
			Type:      quiz.Type,
			CreatorId: quiz.CreatorID,
			CreatedAt: timestamppb.New(quiz.CreatedAt),
			Questions: mapQuestionsToPb(quiz.Questions),
		},
	}, nil
}

// UTILS

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
