package catalog

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/BOURREAUQuentin/game/internal/core/domain"
	"github.com/BOURREAUQuentin/game/internal/core/ports"
	"github.com/machinebox/graphql"
)

type GraphQLCatalogRepository struct {
	client *graphql.Client
}

func NewGraphQLCatalogRepository(url string) ports.CatalogRepository {
	return &GraphQLCatalogRepository{
		client: graphql.NewClient(url),
	}
}

// --- DTOs ---

type CatalogResponse struct {
	Tracks  []TrackDTO  `json:"track_json"`
	Artists []ArtistDTO `json:"artist_json"`
}

type TrackDTO struct {
	TrackID  string `json:"track_id"`
	Title    string `json:"title"`
	ArtistID string `json:"artist_id"`
}

type ArtistDTO struct {
	ArtistID string   `json:"artist_id"`
	Name     string   `json:"name"`
	Genres   []string `json:"genres"`
}

// --- Implementation ---

func (r *GraphQLCatalogRepository) GetQuestionsByGenre(genre string, limit int) ([]domain.Question, error) {
	data, err := r.fetchAllData()
	if err != nil {
		return nil, err
	}

	artistsMap := make(map[string]ArtistDTO)
	allArtistNames := []string{}
	for _, a := range data.Artists {
		artistsMap[a.ArtistID] = a
		allArtistNames = append(allArtistNames, a.Name)
	}

	var validTracks []TrackDTO
	for _, t := range data.Tracks {
		artist, found := artistsMap[t.ArtistID]
		if found && containsGenre(artist.Genres, genre) {
			validTracks = append(validTracks, t)
		}
	}

	return r.processQuestions(validTracks, artistsMap, allArtistNames, limit)
}

func (r *GraphQLCatalogRepository) GetRandomQuestions(limit int) ([]domain.Question, error) {
	data, err := r.fetchAllData()
	if err != nil {
		return nil, err
	}

	artistsMap := make(map[string]ArtistDTO)
	allArtistNames := []string{}
	for _, a := range data.Artists {
		artistsMap[a.ArtistID] = a
		allArtistNames = append(allArtistNames, a.Name)
	}

	return r.processQuestions(data.Tracks, artistsMap, allArtistNames, limit)
}

func (r *GraphQLCatalogRepository) GetQuestionsByArtists(artists []string, limit int) ([]domain.Question, error) {
	// 1. Fetch specific tracks via GraphQL
	req := graphql.NewRequest(`
		query ($artists: [String]!, $limit: Int!) {
			questionsByArtists(artists: $artists, limit: $limit) {
				track_id
				title
				artist_id
			}
		}
	`)
	req.Var("artists", artists)
	req.Var("limit", limit)

	var resp struct {
		Tracks []TrackDTO `json:"questionsByArtists"`
	}

	// Execute GraphQL request
	if err := r.client.Run(context.Background(), req, &resp); err != nil {
		return nil, fmt.Errorf("failed to fetch questions by artists: %w", err)
	}

	// 2. We still need all artists to generate incorrect choices (distractors)
	// Optimization: In a real world scenario, we might fetch only a subset or cache this.
	data, err := r.fetchAllData()
	if err != nil {
		return nil, err
	}

	artistsMap := make(map[string]ArtistDTO)
	allArtistNames := []string{}
	for _, a := range data.Artists {
		artistsMap[a.ArtistID] = a
		allArtistNames = append(allArtistNames, a.Name)
	}

	return r.processQuestions(resp.Tracks, artistsMap, allArtistNames, limit)
}

// --- Helpers ---

func (r *GraphQLCatalogRepository) processQuestions(
	tracks []TrackDTO,
	artistsMap map[string]ArtistDTO,
	allArtistNames []string,
	limit int,
) ([]domain.Question, error) {
	rand.Seed(time.Now().UnixNano())

	shuffledTracks := make([]TrackDTO, len(tracks))
	copy(shuffledTracks, tracks)
	rand.Shuffle(len(shuffledTracks), func(i, j int) { shuffledTracks[i], shuffledTracks[j] = shuffledTracks[j], shuffledTracks[i] })

	count := limit
	if len(shuffledTracks) < count {
		count = len(shuffledTracks)
	}
	selectedTracks := shuffledTracks[:count]

	var questions []domain.Question
	for _, t := range selectedTracks {
		artist, ok := artistsMap[t.ArtistID]
		if !ok {
			continue // Skip if artist details not found
		}

		choices, correctIdx := generateChoices(artist.Name, allArtistNames)

		q := domain.Question{
			QuestionID:       t.TrackID,
			Text:             fmt.Sprintf("Qui est l'interprète du titre \"%s\" ?", t.Title),
			Choices:          choices,
			CorrectAnswerIdx: int32(correctIdx),
		}
		questions = append(questions, q)
	}

	return questions, nil
}

func (r *GraphQLCatalogRepository) fetchAllData() (*CatalogResponse, error) {
	req := graphql.NewRequest(`
		query {
			track_json {
				track_id
				title
				artist_id
			}
			artist_json {
				artist_id
				name
				genres
			}
		}
	`)

	var resp CatalogResponse
	if err := r.client.Run(context.Background(), req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func containsGenre(genres []string, target string) bool {
	for _, g := range genres {
		if g == target {
			return true
		}
	}
	return false
}

func generateChoices(correctAnswer string, allArtists []string) ([]string, int) {
	choices := []string{correctAnswer}

	// Create a local copy to shuffle for distractors
	shuffledAll := make([]string, len(allArtists))
	copy(shuffledAll, allArtists)
	rand.Shuffle(len(shuffledAll), func(i, j int) { shuffledAll[i], shuffledAll[j] = shuffledAll[j], shuffledAll[i] })

	for _, name := range shuffledAll {
		if name != correctAnswer && len(choices) < 4 {
			choices = append(choices, name)
		}
	}

	rand.Shuffle(len(choices), func(i, j int) { choices[i], choices[j] = choices[j], choices[i] })

	var correctIndex int
	for i, choice := range choices {
		if choice == correctAnswer {
			correctIndex = i
			break
		}
	}

	return choices, correctIndex
}
