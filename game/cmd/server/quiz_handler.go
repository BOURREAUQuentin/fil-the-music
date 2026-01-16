package server

import (
	"context"
	"fmt"
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
	questions, err := s.CatalogRepo.GetRandomQuestions(10)
	if err != nil {
		return nil, err
	}

	quiz := &domain.Quiz{
		Title:     "Quiz Aléatoire",
		Type:      "random",
		CreatorID: req.UserId,
		CreatedAt: time.Now(),
		Questions: questions,
	}

	err = s.QuizRepo.CreateQuiz(ctx, quiz)
	if err != nil {
		return nil, err
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
	questions, err := s.CatalogRepo.GetQuestionsByGenre(req.Genre, 10)
	if err != nil {
		return nil, err
	}

	quiz := &domain.Quiz{
		Title:     "Quiz " + req.Genre,
		Type:      "genre",
		CreatorID: req.UserId,
		CreatedAt: time.Now(),
		Questions: questions,
	}

	err = s.QuizRepo.CreateQuiz(ctx, quiz)
	if err != nil {
		return nil, err
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
	// 1. Get User Favorites
	favorites, err := s.UserRepo.GetUserFavorites(ctx, req.UserId)
	fmt.Printf("[DEBUG] UserID: %s, Favorites: %v, Error: %v\n", req.UserId, favorites, err)

	var questions []domain.Question
	quizType := "for_you"
	quizTitle := "Quiz For You"

	// 2. Cold Start Management: Fallback to "Pop" if no favorites or error
	if err != nil || len(favorites) == 0 {
		fmt.Printf("ForYou Fallback triggered for user %s. Cause: err=%v, count=%d\n", req.UserId, err, len(favorites))

		// Fallback to Pop genre
		questions, err = s.CatalogRepo.GetQuestionsByGenre("Electro", 10)
		quizTitle = "Quiz Electro (Recommandé)"
		quizType = "for_you_fallback"
	} else {
		// 3. Normal Flow: Get questions by artists
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

	// 4. Create and Save Quiz
	quiz := &domain.Quiz{
		Title:     quizTitle,
		Type:      quizType,
		CreatorID: req.UserId,
		CreatedAt: time.Now(),
		Questions: questions,
	}

	err = s.QuizRepo.CreateQuiz(ctx, quiz)
	if err != nil {
		return nil, err
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
			QuestionId:         q.QuestionID,
			Text:               q.Text,
			Choices:            q.Choices,
			CorrectAnswerIndex: q.CorrectAnswerIdx,
		})
	}
	return pbQs
}
