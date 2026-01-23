package server

import (
	"context"
	"fmt"
	"time"

	pb "github.com/BOURREAUQuentin/game/api/proto/v1"
	"github.com/BOURREAUQuentin/game/internal/core/domain"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *GameServer) GetQuizzes(ctx context.Context, req *pb.GetQuizzesRequest) (*pb.GetQuizzesResponse, error) {
	quizzesDB, err := s.QuizRepo.FindAllQuizzes(ctx)
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
			CreatedAt: timestamppb.New(q.CreatedAt),
			Questions: mapQuestionsToPb(q.Questions),
		})
	}

	return &pb.GetQuizzesResponse{Quizzes: pbQuizzes}, nil
}

func (s *GameServer) CreateQuiz(ctx context.Context, req *pb.CreateQuizRequest) (*pb.CreateQuizResponse, error) {
	now := time.Now()
	var questions []domain.Question
	for _, q := range req.Questions {
		qID := q.QuestionId
		if qID == "" {
			qID = primitive.NewObjectID().Hex()
		}
		questions = append(questions, domain.Question{
			QuestionID:       qID,
			Text:             q.Text,
			Choices:          q.Choices,
			CorrectAnswerIdx: q.CorrectAnswerIndex,
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

func (s *GameServer) StartRandomQuiz(ctx context.Context, req *pb.StartRandomQuizRequest) (*pb.StartQuizResponse, error) {
	// 1. Get user ID from gRPC metadata
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.DataLoss, "metadata is not provided")
	}
	userIDValues := md.Get("x-user-id")
	if len(userIDValues) == 0 {
		return nil, status.Errorf(codes.Unauthenticated, "user ID is not provided in metadata")
	}
	userID := userIDValues[0]

	// 2. Get questions
	questions, err := s.CatalogRepo.GetRandomQuestions(10)
	if err != nil {
		return nil, err
	}

	// 3. Create quiz with authenticated user's ID
	quiz := &domain.Quiz{
		Title:     "Quiz Aléatoire",
		Type:      "random",
		CreatorID: userID,
		CreatedAt: time.Now(),
		Questions: questions,
	}

	err = s.QuizRepo.CreateQuiz(ctx, quiz)
	if err != nil {
		return nil, err
	}

	// 4. Create Session (Join automatically)
	s.cleanupOldSession(ctx, userID)

	session := &domain.Session{
		UserId:       userID,
		QuizId:       quiz.ID,
		JoinedAt:     time.Now(),
		LastActivity: time.Now(),
	}
	if err := s.SessionRepo.CreateSession(ctx, session); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create session: %v", err)
	}

	return &pb.StartQuizResponse{
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

func (s *GameServer) StartGenreQuiz(ctx context.Context, req *pb.StartGenreQuizRequest) (*pb.StartQuizResponse, error) {
	// 1. Get user ID from gRPC metadata
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.DataLoss, "metadata is not provided")
	}
	userIDValues := md.Get("x-user-id")
	if len(userIDValues) == 0 {
		return nil, status.Errorf(codes.Unauthenticated, "user ID is not provided in metadata")
	}
	userID := userIDValues[0]

	// 2. Get questions
	questions, err := s.CatalogRepo.GetQuestionsByGenre(req.Genre, 10)
	if err != nil {
		return nil, err
	}

	// 3. Create quiz with authenticated user's ID
	quiz := &domain.Quiz{
		Title:     "Quiz " + req.Genre,
		Type:      "genre",
		CreatorID: userID,
		CreatedAt: time.Now(),
		Questions: questions,
	}

	err = s.QuizRepo.CreateQuiz(ctx, quiz)
	if err != nil {
		return nil, err
	}

	// 4. Create Session (Join automatically)
	s.cleanupOldSession(ctx, userID)

	session := &domain.Session{
		UserId:       userID,
		QuizId:       quiz.ID,
		JoinedAt:     time.Now(),
		LastActivity: time.Now(),
	}
	if err := s.SessionRepo.CreateSession(ctx, session); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create session: %v", err)
	}

	return &pb.StartQuizResponse{
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

func (s *GameServer) StartForYouQuiz(ctx context.Context, req *pb.StartForYouQuizRequest) (*pb.StartQuizResponse, error) {
	// 1. Get user ID from gRPC metadata
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.DataLoss, "metadata is not provided")
	}
	userIDValues := md.Get("x-user-id")
	if len(userIDValues) == 0 {
		return nil, status.Errorf(codes.Unauthenticated, "user ID is not provided in metadata")
	}
	userID := userIDValues[0]

	// 2. Get User Favorites using the authenticated ID
	favorites, err := s.UserRepo.GetUserFavorites(ctx, userID)
	fmt.Printf("[DEBUG] UserID: %s, Favorites: %v, Error: %v\n", userID, favorites, err)

	var questions []domain.Question
	quizType := "for_you"
	quizTitle := "Quiz For You"

	// 3. Cold Start Management: Fallback to "Electro" if no favorites or error
	if err != nil || len(favorites) == 0 {
		fmt.Printf("ForYou Fallback triggered for user %s. Cause: err=%v, count=%d\n", userID, err, len(favorites))

		// Fallback to Electro genre
		questions, err = s.CatalogRepo.GetQuestionsByGenre("Electro", 10)
		quizTitle = "Quiz Electro (Fallback)"
		quizType = "for_you_fallback"
	} else {
		// 4. Normal Flow: Get questions by artists
		questions, err = s.CatalogRepo.GetQuestionsByArtists(favorites, 10)

		// Additional safety: if favorites exist but no questions found (e.g. artists not in catalog)
		if err == nil && len(questions) == 0 {
			questions, err = s.CatalogRepo.GetQuestionsByGenre("Electro", 10)
			quizTitle = "Quiz Electro (Fallback)"
			quizType = "for_you_fallback"
		}
	}

	if err != nil {
		return nil, err
	}

	// 5. Create and Save Quiz
	quiz := &domain.Quiz{
		Title:     quizTitle,
		Type:      quizType,
		CreatorID: userID,
		CreatedAt: time.Now(),
		Questions: questions,
	}

	err = s.QuizRepo.CreateQuiz(ctx, quiz)
	if err != nil {
		return nil, err
	}

	// 6. Create Session (Join automatically)
	s.cleanupOldSession(ctx, userID)

	session := &domain.Session{
		UserId:       userID,
		QuizId:       quiz.ID,
		JoinedAt:     time.Now(),
		LastActivity: time.Now(),
	}
	if err := s.SessionRepo.CreateSession(ctx, session); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create session: %v", err)
	}

	return &pb.StartQuizResponse{
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
			Choices:    q.Choices,
		})
	}
	return pbQs
}
