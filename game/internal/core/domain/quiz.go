package domain

import "time"

type Quiz struct {
	ID        string
	Title     string
	Type      string
	CreatorID string
	Status    string
	CreatedAt time.Time
	Questions []Question
}

type Question struct {
	QuestionID string
	Text       string
}
