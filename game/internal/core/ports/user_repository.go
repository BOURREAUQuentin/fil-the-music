package ports

import (
	"context"

	"github.com/BOURREAUQuentin/game/internal/core/domain"
)

// UserRepository définit l'interface pour interagir avec le service User
type UserRepository interface {
	// GetUserFavorites récupère la liste des artistes favoris d'un utilisateur
	GetUserFavorites(ctx context.Context, userID string) ([]string, error)
	// GetUserById récupère les informations d'un utilisateur par son ID
	GetUserById(ctx context.Context, id string) (*domain.User, error)
}
