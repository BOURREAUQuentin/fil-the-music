package repository

import (
	"context"
	"log"

	"github.com/BOURREAUQuentin/game/internal/core/domain"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type MongoRepository struct {
	collection *mongo.Collection
}

func NewMongoRepository(db *mongo.Database) *MongoRepository {
	return &MongoRepository{
		collection: db.Collection("quiz"),
	}
}

func (r *MongoRepository) FindAll(ctx context.Context) ([]domain.Quiz, error) {
	filter := bson.M{}

	// Find collection
	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var quizzes []domain.Quiz

	// Retrieve quizzes
	for cursor.Next(ctx) {
		var doc QuizMongoDoc
		if err := cursor.Decode(&doc); err != nil {
			log.Printf("Erreur de décodage d'un document: %v", err)
			continue
		}

		// Add cleaned data
		quizzes = append(quizzes, doc.ToDomain())
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return quizzes, nil
}
