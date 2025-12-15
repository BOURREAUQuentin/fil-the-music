package repository

import (
	"context"
	"log"

	"github.com/BOURREAUQuentin/game/internal/core/domain"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type MongoQuizRepository struct {
	quizCollection *mongo.Collection
}

func NewMongoQuizRepository(db *mongo.Database) *MongoQuizRepository {
	return &MongoQuizRepository{
		quizCollection: db.Collection("quiz"),
	}
}

func (r *MongoQuizRepository) FindAllQuizzes(ctx context.Context) ([]domain.Quiz, error) {
	filter := bson.M{}

	// Find collection
	cursor, err := r.quizCollection.Find(ctx, filter)
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

func (r *MongoQuizRepository) FindQuizByID(ctx context.Context, id string) (*domain.Quiz, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	filter := bson.M{"_id": objID}
	var doc QuizMongoDoc
	err = r.quizCollection.FindOne(ctx, filter).Decode(&doc)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil // Quiz not found
		}
		return nil, err
	}

	quiz := doc.ToDomain()
	return &quiz, nil
}
