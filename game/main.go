package main

import (
	"context"
	"log"
	"net"
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

	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://root:example@localhost:27017/"))
	if err != nil {
		log.Fatal(err)
	}

	db := client.Database("game")

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
