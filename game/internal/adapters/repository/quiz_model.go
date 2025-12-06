package repository

import (
	"time"

	"github.com/BOURREAUQuentin/game/internal/core/domain"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type QuizMongoDoc struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	Title     string             `bson:"title"`
	Type      string             `bson:"type"`
	CreatorID string             `bson:"creator_id"`
	Status    string             `bson:"status"`
	CreatedAt time.Time          `bson:"created_at"`
	Questions []QuestionMongoDoc `bson:"questions"`
}

type QuestionMongoDoc struct {
	QuestionID string `bson:"question_id"`
	Text       string `bson:"text"`
}

// ToDomain convert BSON document into cleaned data
func (d *QuizMongoDoc) ToDomain() domain.Quiz {
	questions := make([]domain.Question, len(d.Questions))
	for i, q := range d.Questions {
		questions[i] = domain.Question{
			QuestionID: q.QuestionID,
			Text:       q.Text,
		}
	}

	return domain.Quiz{
		ID:        d.ID.Hex(),
		Title:     d.Title,
		Type:      d.Type,
		CreatorID: d.CreatorID,
		Status:    d.Status,
		CreatedAt: d.CreatedAt,
		Questions: questions,
	}
}
