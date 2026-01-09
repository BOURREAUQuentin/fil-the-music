package repository

import (
	"context"
	"log"
	"time"

	"github.com/BOURREAUQuentin/game/internal/core/domain"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type MongoSessionRepository struct {
	sessionCollection *mongo.Collection
}

func NewMongoSessionRepository(db *mongo.Database) *MongoSessionRepository {
	return &MongoSessionRepository{
		sessionCollection: db.Collection("active_session"),
	}
}

func (m MongoSessionRepository) GetSessions(ctx context.Context) ([]domain.Session, error) {
	filter := bson.M{}

	// Find collection
	cursor, err := m.sessionCollection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var sessions []domain.Session

	// Retrieve sessions
	for cursor.Next(ctx) {
		var doc SessionMongoDoc
		if err := cursor.Decode(&doc); err != nil {
			log.Printf("Erreur de décodage d'un document: %v", err)
			continue
		}

		// Add cleaned data
		sessions = append(sessions, doc.ToDomain())
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return sessions, nil
}

func (m MongoSessionRepository) GetSessionByUserId(ctx context.Context, userId string) (*domain.Session, error) {
	filter := bson.M{"user_id": userId}
	var doc SessionMongoDoc
	err := m.sessionCollection.FindOne(ctx, filter).Decode(&doc)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil // Session not found
		}
		return nil, err
	}

	session := doc.ToDomain()
	return &session, nil
}

func (m MongoSessionRepository) CreateSession(ctx context.Context, session *domain.Session) error {
	doc := SessionMongoDoc{
		UserId:       session.UserId,
		QuizId:       session.QuizId,
		JoinedAt:     session.JoinedAt,
		LastActivity: session.LastActivity,
	}

	result, err := m.sessionCollection.InsertOne(ctx, doc)
	if err != nil {
		return err
	}

	if oid, ok := result.InsertedID.(primitive.ObjectID); ok {
		session.ID = oid.Hex()
	}

	return nil
}

func (m MongoSessionRepository) DeleteSession(ctx context.Context, session *domain.Session) error {
	filter := bson.M{
		"user_id": session.UserId,
		"quiz_id": session.QuizId,
	}

	_, err := m.sessionCollection.DeleteOne(ctx, filter)
	return err
}

func (m MongoSessionRepository) UpdateActivity(ctx context.Context, sessionId string, lastActivity time.Time) error {
	//TODO implement me
	panic("implement me")
}
