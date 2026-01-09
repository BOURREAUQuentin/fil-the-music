package main

import (
	"context"
	"encoding/json"
	"log"
	"net"
	"os"
	"time"

	pb "github.com/BOURREAUQuentin/game/api/proto/v1"
	"github.com/BOURREAUQuentin/game/cmd/server"
	"github.com/BOURREAUQuentin/game/internal/adapters/repository"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		mongoURI = "mongodb://root:example@localhost:27017/"
	}

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatal(err)
	}

	db := client.Database("game")

	// Seed database if empty
	log.Println("Checking if seeding is required...")
	seed(db)

	quizRepo := repository.NewMongoQuizRepository(db)
	sessionRepo := repository.NewMongoSessionRepository(db)
	userRepo := repository.NewUserRepository()

	gameServer := server.NewGameServer(quizRepo, sessionRepo, userRepo)

	// Launch server
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatal(err)
	}

	s := grpc.NewServer()
	pb.RegisterGameServiceServer(s, gameServer)

	log.Println("grpc server is listening on :50051")
	if err := s.Serve(lis); err != nil {
		log.Fatal(err)
	}
}

func seed(db *mongo.Database) {
	ctx := context.Background()

	// Seed Quizzes
	quizColl := db.Collection("quiz")
	count, _ := quizColl.EstimatedDocumentCount(ctx)
	if count == 0 {
		log.Println("Seeding quizzes...")
		data, err := os.ReadFile("init/quiz.json")
		if err != nil {
			log.Printf("Could not read init/quiz.json: %v", err)
		} else {
			var rawQuizzes []map[string]interface{}
			if err := json.Unmarshal(data, &rawQuizzes); err != nil {
				log.Printf("Could not unmarshal init/quiz.json: %v", err)
			} else {
				var docs []interface{}
				for _, q := range rawQuizzes {
					// Handle extended JSON date format for created_at
					if ca, ok := q["created_at"].(map[string]interface{}); ok {
						if dateStr, ok := ca["$date"].(string); ok {
							parsedTime, err := time.Parse(time.RFC3339, dateStr)
							if err == nil {
								q["created_at"] = parsedTime
							}
						}
					}
					docs = append(docs, q)
				}
				if len(docs) > 0 {
					if _, err := quizColl.InsertMany(ctx, docs); err != nil {
						log.Printf("Could not insert quizzes: %v", err)
					} else {
						log.Println("Quizzes seeded successfully.")
					}
				}
			}
		}
	}

	// Seed Sessions
	sessionColl := db.Collection("active_session")
	count, _ = sessionColl.EstimatedDocumentCount(ctx)
	if count == 0 {
		log.Println("Seeding sessions...")
		data, err := os.ReadFile("init/active_session.json")
		if err != nil {
			log.Printf("Could not read init/active_session.json: %v", err)
		} else {
			var rawSessions []map[string]interface{}
			if err := json.Unmarshal(data, &rawSessions); err != nil {
				log.Printf("Could not unmarshal init/active_session.json: %v", err)
			} else {
				var docs []interface{}
				for _, s := range rawSessions {
					joinedAt, _ := time.Parse(time.RFC3339, s["joined_at"].(string))
					lastActivity, _ := time.Parse(time.RFC3339, s["last_activity"].(string))

					doc := repository.SessionMongoDoc{
						UserId:       s["user_id"].(string),
						QuizId:       s["quiz_id"].(string),
						JoinedAt:     joinedAt,
						LastActivity: lastActivity,
					}
					docs = append(docs, doc)
				}
				if len(docs) > 0 {
					if _, err := sessionColl.InsertMany(ctx, docs); err != nil {
						log.Printf("Could not insert sessions: %v", err)
					} else {
						log.Println("Sessions seeded successfully.")
					}
				}
			}
		}
	}
}
