package domain

import "time"

type Session struct {
	ID           string
	UserId       string
	QuizId       string
	JoinedAt     time.Time
	LastActivity time.Time
}
