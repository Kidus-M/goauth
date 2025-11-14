package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"go_mango/controllers"
	"go_mango/data"
	"go_mango/router"
)

func main() {
	// Load .env if present
	_ = godotenv.Load()

	mongoURI := os.Getenv("MONGODB_URI")
	dbName := os.Getenv("MONGODB_DATABASE")
	collectionName := os.Getenv("MONGODB_COLLECTION")
	if mongoURI == "" || dbName == "" || collectionName == "" {
		log.Fatal("MONGODB_URI, MONGODB_DATABASE and MONGODB_COLLECTION must be set (see .env.example)")
	}

	// Connect to MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := data.NewMongoClient(ctx, mongoURI)
	if err != nil {
		log.Fatalf("failed to create mongo client: %v", err)
	}

	// Close client on exit
	defer func() {
		_ = client.Disconnect(context.Background())
	}()

	collection := client.Database(dbName).Collection(collectionName)
	taskService := data.NewTaskService(collection)
	taskController := controllers.NewTaskController(taskService)

	r := router.SetupRouter(taskController)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	addr := fmt.Sprintf(":%s", port)
	if err := r.Run(addr); err != nil {
		log.Fatalf("failed to run server: %v", err)
	}
}
