package ports

import "github.com/BOURREAUQuentin/game/internal/core/domain"

type CatalogRepository interface {
	// Récupère des questions filtrées par genre
	GetQuestionsByGenre(genre string, limit int) ([]domain.Question, error)

	// Récupère des questions aléatoires (pour le mode "For You" en attendant la reco)
	GetRandomQuestions(limit int) ([]domain.Question, error)
}
