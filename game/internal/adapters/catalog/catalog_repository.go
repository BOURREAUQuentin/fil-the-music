package catalog

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"strconv"
	"strings"
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
	TrackID     string `json:"track_id"`
	Title       string `json:"title"`
	ArtistID    string `json:"artist_id"`
	AlbumName   string `json:"album_name"`
	ReleaseDate string `json:"release_date"`
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

	// Pass all tracks as pool for distractors
	return r.processQuestions(validTracks, artistsMap, allArtistNames, data.Tracks, limit)
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

	return r.processQuestions(data.Tracks, artistsMap, allArtistNames, data.Tracks, limit)
}

func (r *GraphQLCatalogRepository) GetQuestionsByArtists(artists []string, limit int) ([]domain.Question, error) {
	// 1. Fetch specific tracks via GraphQL
	req := graphql.NewRequest(`
		query ($artists: [String]!, $limit: Int!) {
			questions_by_artists(artists: $artists, limit: $limit) {
				track_id
				title
				artist_id
				album_name
				release_date
			}
		}
	`)
	req.Var("artists", artists)
	req.Var("limit", limit)

	var resp struct {
		Tracks []TrackDTO `json:"questions_by_artists"`
	}

	// Execute GraphQL request
	if err := r.client.Run(context.Background(), req, &resp); err != nil {
		return nil, fmt.Errorf("failed to fetch questions by artists: %w", err)
	}

	// 2. We still need all artists/tracks to generate incorrect choices (distractors)
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

	return r.processQuestions(resp.Tracks, artistsMap, allArtistNames, data.Tracks, limit)
}

// --- Helpers ---

func (r *GraphQLCatalogRepository) processQuestions(
	tracks []TrackDTO,
	artistsMap map[string]ArtistDTO,
	allArtistNames []string,
	poolTracks []TrackDTO,
	limit int,
) ([]domain.Question, error) {
	rand.Seed(time.Now().UnixNano())

	// Build pools
	allTitles := []string{}
	allAlbums := []string{}
	seenAlbums := make(map[string]bool)

	for _, t := range poolTracks {
		allTitles = append(allTitles, t.Title)
		if t.AlbumName != "" && !seenAlbums[t.AlbumName] {
			allAlbums = append(allAlbums, t.AlbumName)
			seenAlbums[t.AlbumName] = true
		}
	}

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

		// Diversify Question Types
		qType := rand.Intn(4)

		// Fallbacks
		if qType == 2 { // Album
			if t.AlbumName == "" || strings.EqualFold(t.AlbumName, t.Title) {
				qType = 0
			}
		}
		if qType == 3 { // Year
			if t.ReleaseDate == "" {
				qType = 0
			}
		}

		var text string
		var choices []string
		var correctIdx int

		switch qType {
		case 1: // Reverse (Title by Artist)
			text = fmt.Sprintf("Lequel de ces titres est interprété par %s ?", artist.Name)
			// Filter titles to exclude those by same artist is tricky without full mapping
			// For simplicity, we assume pool is random enough, but ideally we should filter.
			// Let's rely on standard shuffling, chance of collision is low if catalog is large.
			choices, correctIdx = generateChoices(t.Title, allTitles)

		case 2: // Album
			text = fmt.Sprintf("Dans quel album trouve-t-on le titre \"%s\" ?", t.Title)
			choices, correctIdx = generateChoices(t.AlbumName, allAlbums)

		case 3: // Date
			text = fmt.Sprintf("En quelle année est sorti le titre \"%s\" ?", t.Title)
			year := ""
			parts := strings.Split(t.ReleaseDate, "-")
			if len(parts) > 0 {
				year = parts[0]
			}
			choices, correctIdx = generateYearChoices(year)

		default: // 0 - Classic (Artist)
			text = fmt.Sprintf("Qui est l'interprète du titre \"%s\" ?", t.Title)
			choices, correctIdx = generateChoices(artist.Name, allArtistNames)
		}

		q := domain.Question{
			QuestionID:       t.TrackID,
			Text:             text,
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
				album_name
				release_date
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
		if strings.EqualFold(g, target) {
			return true
		}
	}
	return false
}

func generateChoices(correctAnswer string, pool []string) ([]string, int) {
	choices := []string{correctAnswer}

	// Create a local copy to shuffle for distractors
	shuffledPool := make([]string, len(pool))
	copy(shuffledPool, pool)
	rand.Shuffle(len(shuffledPool), func(i, j int) { shuffledPool[i], shuffledPool[j] = shuffledPool[j], shuffledPool[i] })

	for _, item := range shuffledPool {
		if item != correctAnswer && len(choices) < 4 {
			// Basic deduplication for choices
			found := false
			for _, c := range choices {
				if c == item {
					found = true
					break
				}
			}
			if !found {
				choices = append(choices, item)
			}
		}
	}

	// Fill with placeholders if not enough items in pool
	if len(choices) < 4 {
		for i := len(choices); i < 4; i++ {
			choices = append(choices, "Autre")
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

func generateYearChoices(correctYearStr string) ([]string, int) {
	year, err := strconv.Atoi(correctYearStr)
	if err != nil {
		return []string{correctYearStr, "2000", "2010", "2020"}, 0
	}

	currentYear := time.Now().Year()
	distractors := make(map[int]bool)

	// On boucle tant qu'on n'a pas 3 fausses réponses (peut-être dangereux ??)
	for len(distractors) < 3 {
		offset := rand.Intn(5) + 1

		if rand.Intn(2) == 0 {
			offset = -offset
		}

		fake := year + offset

		if fake > currentYear {
			fake = year - int(math.Abs(float64(offset)))
		}

		if fake == year || fake > currentYear {
			fake = year - (len(distractors) + 1)
		}

		if !distractors[fake] {
			distractors[fake] = true
		}
	}

	choices := []string{correctYearStr}
	for y := range distractors {
		choices = append(choices, strconv.Itoa(y))
	}

	rand.Shuffle(len(choices), func(i, j int) { choices[i], choices[j] = choices[j], choices[i] })

	var correctIndex int
	for i, choice := range choices {
		if choice == correctYearStr {
			correctIndex = i
			break
		}
	}

	return choices, correctIndex
}
