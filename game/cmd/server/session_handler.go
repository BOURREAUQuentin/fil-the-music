package server

import (
	"context"
	"time"

	pb "github.com/BOURREAUQuentin/game/api/proto/v1"
	"github.com/BOURREAUQuentin/game/internal/core/domain"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *GameServer) JoinQuiz(ctx context.Context, req *pb.JoinQuizRequest) (*pb.JoinQuizResponse, error) {
	// 1. Get user ID from gRPC metadata (trusted from gateway)
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.DataLoss, "metadata is not provided")
	}

	userIDValues := md.Get("x-user-id")
	if len(userIDValues) == 0 {
		return nil, status.Errorf(codes.Unauthenticated, "user ID is not provided in metadata")
	}
	userID := userIDValues[0]

	// 2. Validate quiz_id format
	_, err := primitive.ObjectIDFromHex(req.QuizId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid quiz ID format: %v", err)
	}

	// 3. Verify quiz existence
	quiz, err := s.QuizRepo.FindQuizByID(ctx, req.QuizId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to retrieve quiz: %v", err)
	}
	if quiz == nil {
		return nil, status.Errorf(codes.NotFound, "quiz with ID %s not found", req.QuizId)
	}

	// 4. Check if user already has an active session
	existingSession, err := s.SessionRepo.GetSessionByUserId(ctx, userID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to check existing session: %v", err)
	}
	if existingSession != nil {
		return nil, status.Errorf(codes.AlreadyExists, "user %s already has an active session", userID)
	}

	// 5. Create the session
	now := time.Now()
	session := &domain.Session{
		UserId:       userID, // Use ID from metadata
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

func (s *GameServer) QuitQuiz(ctx context.Context, req *pb.QuitQuizRequest) (*pb.QuitQuizResponse, error) {
	// 1. Get user ID from gRPC metadata (trusted from gateway)
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.DataLoss, "metadata is not provided")
	}
	userIDValues := md.Get("x-user-id")
	if len(userIDValues) == 0 {
		return nil, status.Errorf(codes.Unauthenticated, "user ID is not provided in metadata")
	}
	userID := userIDValues[0]

	// 2. Validate quiz_id format
	_, err := primitive.ObjectIDFromHex(req.QuizId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "failed to parse quiz ID: %v", err)
	}

	// 3. Verify quiz existence (optional but good practice)
	quiz, err := s.QuizRepo.FindQuizByID(ctx, req.QuizId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to retrieve quiz: %v", err)
	}
	if quiz == nil {
		return nil, status.Errorf(codes.NotFound, "quiz with ID %s not found", req.QuizId)
	}

	// 4. Delete the session for the authenticated user
	session := &domain.Session{
		UserId: userID,
		QuizId: req.QuizId,
	}

	err = s.SessionRepo.DeleteSession(ctx, session)
	if err != nil {
		// Consider the case where the session does not exist as a success for idempotency
		if err.Error() == "mongo: no documents in result" {
			return &pb.QuitQuizResponse{Success: true}, nil
		}
		return nil, status.Errorf(codes.Internal, "failed to delete session: %v", err)
	}

	// Delete quiz if it was auto-generated
	if quiz.Type == "random" || quiz.Type == "genre" || quiz.Type == "for_you" || quiz.Type == "for_you_fallback" {
		_ = s.QuizRepo.DeleteQuiz(ctx, quiz.ID)
	}

	return &pb.QuitQuizResponse{
		Success: true,
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

func (s *GameServer) GetActiveQuiz(ctx context.Context, req *pb.GetActiveQuizRequest) (*pb.GetActiveQuizResponse, error) {
	// 1. Get user ID from gRPC metadata (trusted from gateway)
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.DataLoss, "metadata is not provided")
	}
	userIDValues := md.Get("x-user-id")
	if len(userIDValues) == 0 {
		return nil, status.Errorf(codes.Unauthenticated, "user ID is not provided in metadata")
	}
	userID := userIDValues[0]

	// 2. Find session using the authenticated user ID
	session, err := s.SessionRepo.GetSessionByUserId(ctx, userID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to retrieve session: %v", err)
	}
	if session == nil {
		return nil, status.Errorf(codes.NotFound, "no active session found for user %s", userID)
	}

	// 3. Find the associated quiz
	quiz, err := s.QuizRepo.FindQuizByID(ctx, session.QuizId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to retrieve quiz: %v", err)
	}
	if quiz == nil {
		return nil, status.Errorf(codes.NotFound, "quiz with ID %s not found", session.QuizId)
	}

	return &pb.GetActiveQuizResponse{
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

func (s *GameServer) AnswerQuestions(ctx context.Context, req *pb.AnswerQuestionsRequest) (*pb.AnswerQuestionsResponse, error) {
	// 1. Get user ID from gRPC metadata (trusted from gateway)
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.DataLoss, "metadata is not provided")
	}
	userIDValues := md.Get("x-user-id")
	if len(userIDValues) == 0 {
		return nil, status.Errorf(codes.Unauthenticated, "user ID is not provided in metadata")
	}
	userID := userIDValues[0]

	// 2. Find the user's active session
	session, err := s.SessionRepo.GetSessionByUserId(ctx, userID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to retrieve session: %v", err)
	}
	if session == nil {
		return nil, status.Errorf(codes.NotFound, "no active session found for user %s", userID)
	}

	// 3. Find the quiz associated with the session
	quiz, err := s.QuizRepo.FindQuizByID(ctx, session.QuizId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to retrieve quiz: %v", err)
	}
	if quiz == nil {
		return nil, status.Errorf(codes.NotFound, "quiz with ID %s not found", session.QuizId)
	}

	if len(req.Answers) != len(quiz.Questions) {
		return nil, status.Errorf(codes.InvalidArgument, "number of answers (%d) does not match number of questions (%d)", len(req.Answers), len(quiz.Questions))
	}

	// 4. Calculate score
	var score int32
	var results []*pb.QuestionResult

	for i, question := range quiz.Questions {
		userAnswerIdx := req.Answers[i]
		isCorrect := userAnswerIdx == question.CorrectAnswerIdx

		if isCorrect {
			score++
		}

		results = append(results, &pb.QuestionResult{
			QuestionId:         question.QuestionID,
			IsCorrect:          isCorrect,
			CorrectAnswerIndex: question.CorrectAnswerIdx,
			UserAnswerIndex:    userAnswerIdx,
		})
	}

	// 5. Delete session after quiz completion
	s.SessionRepo.DeleteSession(ctx, session)

	// 6. Delete quiz if it was auto-generated
	if quiz.Type == "random" || quiz.Type == "genre" || quiz.Type == "for_you" || quiz.Type == "for_you_fallback" {
		_ = s.QuizRepo.DeleteQuiz(ctx, quiz.ID)
	}

	return &pb.AnswerQuestionsResponse{
		Score:   score,
		Results: results,
	}, nil
}
