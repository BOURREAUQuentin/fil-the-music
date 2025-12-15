package server

import (
	"context"
	"time"

	pb "github.com/BOURREAUQuentin/game/api/proto/v1"
	"github.com/BOURREAUQuentin/game/internal/core/domain"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *GameServer) JoinQuiz(ctx context.Context, req *pb.JoinQuizRequest) (*pb.JoinQuizResponse, error) {
	// Validate quiz_id format
	_, err := primitive.ObjectIDFromHex(req.QuizId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to parse quiz ID: %v", err)
	}

	user, err := s.UserRepo.GetUserById(ctx, req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to retrieve user: %v", err)
	}

	// Verify quiz_id existence
	quiz, err := s.QuizRepo.FindQuizByID(ctx, req.QuizId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to retrieve quiz: %v", err)
	}

	if quiz == nil {
		return nil, status.Errorf(codes.NotFound, "quiz with ID %s not found", req.QuizId)
	}
	if user == nil {
		return nil, status.Errorf(codes.NotFound, "user with ID %s not found", req.UserId)
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
