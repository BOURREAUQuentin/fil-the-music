package catalog

import (
	"context"
	"fmt"
	"github.com/BOURREAUQuentin/game/internal/core/domain" // Vérifie ton import
	"github.com/BOURREAUQuentin/game/internal/core/ports"
	"math/rand"
	"time"

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

// --- DTOs pour lire le JSON qui vient de Python ---
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

// --- Implémentation ---

func (r *GraphQLCatalogRepository) GetQuestionsByGenre(genre string, limit int) ([]domain.Question, error) {
	// 1. Récupérer tout le catalogue
	data, err := r.fetchAllData()
	if err != nil {
		return nil, err
	}

	// 2. Indexer les artistes pour un accès rapide (ID -> Artiste)
	artistsMap := make(map[string]ArtistDTO)
	allArtistNames := []string{} // Sert pour générer les mauvaises réponses
	for _, a := range data.Artists {
		artistsMap[a.ArtistID] = a
		allArtistNames = append(allArtistNames, a.Name)
	}

	// 3. Filtrer les tracks par genre
	var validTracks []TrackDTO
	for _, t := range data.Tracks {
		artist, found := artistsMap[t.ArtistID]
		if found && containsGenre(artist.Genres, genre) {
			validTracks = append(validTracks, t)
		}
	}

	return r.processQuestions(validTracks, artistsMap, allArtistNames, limit)
}

// Fonction pour le "For You" (identique mais sans filtre genre)
func (r *GraphQLCatalogRepository) GetRandomQuestions(limit int) ([]domain.Question, error) {
	// 1. Récupérer tout le catalogue
	data, err := r.fetchAllData()
	if err != nil {
		return nil, err
	}

	// 2. Indexer les artistes
	artistsMap := make(map[string]ArtistDTO)
	allArtistNames := []string{}
	for _, a := range data.Artists {
		artistsMap[a.ArtistID] = a
		allArtistNames = append(allArtistNames, a.Name)
	}

	// 3. Tous les tracks sont valides
	// On copie juste les tracks, ou on utilise directement data.Tracks
	// Ici data.Tracks est []TrackDTO
	validTracks := data.Tracks

	return r.processQuestions(validTracks, artistsMap, allArtistNames, limit)
}

// Méthode commune pour mélanger, limiter et convertir en questions
func (r *GraphQLCatalogRepository) processQuestions(
	tracks []TrackDTO,
	artistsMap map[string]ArtistDTO,
	allArtistNames []string,
	limit int,
) ([]domain.Question, error) {

	// 4. Mélanger les tracks pour ne pas avoir toujours les mêmes
	// Note: rand.Seed est global, idéalement à faire une fois au main ou init,
	// mais ici on le laisse pour l'instant (bien que deprecated en Go 1.20+)
	rand.Seed(time.Now().UnixNano())

	// Copie pour ne pas modifier l'original si nécessaire (ici tracks est déjà une copie/slice locale)
	shuffledTracks := make([]TrackDTO, len(tracks))
	copy(shuffledTracks, tracks)

	rand.Shuffle(len(shuffledTracks), func(i, j int) { shuffledTracks[i], shuffledTracks[j] = shuffledTracks[j], shuffledTracks[i] })

	// Limiter au nombre demandé
	count := limit
	if len(shuffledTracks) < count {
		count = len(shuffledTracks)
	}
	selectedTracks := shuffledTracks[:count]

	// 5. Construire les questions du domaine
	var questions []domain.Question

	for _, t := range selectedTracks {
		correctArtistName := artistsMap[t.ArtistID].Name

		// Générer 3 mauvaises réponses + la bonne
		choices, correctIdx := generateChoices(correctArtistName, allArtistNames)

		q := domain.Question{
			QuestionID:       t.TrackID, // On utilise l'ID du track comme ID de question
			Text:             fmt.Sprintf("Qui est l'interprète du titre \"%s\" ?", t.Title),
			Choices:          choices,
			CorrectAnswerIdx: int32(correctIdx),
		}
		questions = append(questions, q)
	}

	return questions, nil
}

// --- Helpers Privés ---

func (r *GraphQLCatalogRepository) fetchAllData() (*CatalogResponse, error) {
	// La query GraphQL correspond exactement à tes fonctions Python track_json et artist_json
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
		// Amélioration possible : strings.EqualFold(g, target) pour ignorer la casse
		if g == target {
			return true
		}
	}
	return false
}

// Génère la liste mélangée et retourne l'index de la bonne réponse
func generateChoices(correctAnswer string, allArtists []string) ([]string, int) {
	// 1. Prendre la bonne réponse
	choices := []string{correctAnswer}

	// 2. Prendre 3 mauvaises réponses au hasard
	// On mélange toute la liste d'artistes pour piocher dedans
	shuffledAll := make([]string, len(allArtists))
	copy(shuffledAll, allArtists)
	rand.Shuffle(len(shuffledAll), func(i, j int) { shuffledAll[i], shuffledAll[j] = shuffledAll[j], shuffledAll[i] })

	for _, name := range shuffledAll {
		if name != correctAnswer && len(choices) < 4 {
			choices = append(choices, name)
		}
	}

	// 3. Mélanger les 4 choix finaux pour que la bonne réponse ne soit pas toujours la première
	rand.Shuffle(len(choices), func(i, j int) { choices[i], choices[j] = choices[j], choices[i] })

	// 4. Retrouver l'index de la bonne réponse
	var correctIndex int
	for i, choice := range choices {
		if choice == correctAnswer {
			correctIndex = i
			break
		}
	}

	return choices, correctIndex
}
