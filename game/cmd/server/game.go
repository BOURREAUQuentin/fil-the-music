package server

import (
	"context"
	"time"

	pb "github.com/BOURREAUQuentin/game/api/proto/v1"
	"github.com/BOURREAUQuentin/game/internal/core/domain"
	"github.com/BOURREAUQuentin/game/internal/core/ports"
	"go.mongodb.org/mongo-driver/bson/primitive" // Import primitive
	"google.golang.org/grpc/codes"               // Import codes
	"google.golang.org/grpc/status"              // Import status
	"google.golang.org/protobuf/types/known/timestamppb"
)

type GameServer struct {
	pb.UnimplementedGameServiceServer
	QuizRepo    ports.QuizRepository
	SessionRepo ports.SessionRepository
}

func (s *GameServer) JoinQuiz(ctx context.Context, req *pb.JoinQuizRequest) (*pb.JoinQuizResponse, error) {
	// Validate quiz_id format
	_, err := primitive.ObjectIDFromHex(req.QuizId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to parse quiz ID: %v", err)
	}

	// Verify quiz_id existence
	quiz, err := s.QuizRepo.FindQuizByID(ctx, req.QuizId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to retrieve quiz: %v", err)
	}

	if quiz == nil {
		return nil, status.Errorf(codes.NotFound, "quiz with ID %s not found", req.QuizId)
	}

	now := time.Now()
	session := &domain.Session{
		UserId:       req.UserId,
		QuizId:       req.QuizId,
		JoinedAt:     now,
		LastActivity: now,
	}

	err = s.SessionRepo.CreateSession(ctx, session)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create session: %v", err)
	}

	return &pb.JoinQuizResponse{
		Session: &pb.Session{
			Id:           session.ID,
			UserId:       session.UserId,
			QuizId:       session.QuizId,
			JoinedAt:     timestamppb.New(session.JoinedAt),
			LastActivity: timestamppb.New(session.LastActivity),
		},
	}, nil
}

func (s *GameServer) GetSessions(ctx context.Context, req *pb.GetSessionsRequest) (*pb.GetSessionsResponse, error) {
	sessionsDB, err := s.SessionRepo.GetSessions(ctx)
	if err != nil {
		return nil, err
	}

	// Format datas
	var pbSessions []*pb.Session
	for _, session := range sessionsDB {
		pbSessions = append(pbSessions, &pb.Session{
			Id:           session.ID,
			UserId:       session.UserId,
			QuizId:       session.QuizId,
			JoinedAt:     timestamppb.New(session.JoinedAt),
			LastActivity: timestamppb.New(session.LastActivity),
		})
	}

	return &pb.GetSessionsResponse{Sessions: pbSessions}, nil
}

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
