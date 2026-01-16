package user

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/BOURREAUQuentin/game/internal/core/domain"
	"github.com/BOURREAUQuentin/game/internal/core/ports"
)

type HttpUserRepository struct {
	baseURL string
	client  *http.Client
}

func NewHttpUserRepository(baseURL string) ports.UserRepository {
	return &HttpUserRepository{
		baseURL: baseURL,
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

type UserFavoritesResponse struct {
	FavoriteArtists []string `json:"favoriteArtists"`
}

func (r *HttpUserRepository) GetUserFavorites(ctx context.Context, userID string) ([]string, error) {
	// Construct URL: http://user:3000/users/{userID}/favorites
	url := fmt.Sprintf("%s/users/%s/favorites", r.baseURL, userID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := r.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// If user not found or no favorites, return empty list to allow fallback logic
		if resp.StatusCode == http.StatusNotFound {
			return []string{}, nil
		}
		return nil, fmt.Errorf("user service returned status: %d", resp.StatusCode)
	}

	var userResp UserFavoritesResponse
	if err := json.NewDecoder(resp.Body).Decode(&userResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return userResp.FavoriteArtists, nil
}

func (r *HttpUserRepository) GetUserById(ctx context.Context, id string) (*domain.User, error) {
	// Placeholder implementation as per previous state, ensuring it satisfies the interface
	return &domain.User{
		ID:       id,
		Username: "RemoteUser",
		Role:     domain.RoleUser,
	}, nil
}
