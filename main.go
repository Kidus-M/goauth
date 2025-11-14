package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"authgo/controllers"
	"authgo/data"
	"authgo/middleware"
	"authgo/router"

	"github.com/joho/godotenv"
)

func main() {
	// load env
	_ = godotenv.Load()

	uri := os.Getenv("MONGODB_URI")
	dbName := os.Getenv("MONGODB_DATABASE")
	taskCollName := os.Getenv("MONGODB_TASK_COLLECTION")
	userCollName := os.Getenv("MONGODB_USER_COLLECTION")
	jwtSecret := os.Getenv("JWT_SECRET")

	if uri == "" || dbName == "" || taskCollName == "" || userCollName == "" || jwtSecret == "" {
		log.Fatal("MONGODB_URI, MONGODB_DATABASE, collections and JWT_SECRET must be set (see .env.example)")
	}

	// connect to mongo
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	client, err := data.NewMongoClient(ctx, uri)
	if err != nil {
		log.Fatalf("failed to connect to mongodb: %v", err)
	}
	defer func() {
		_ = client.Disconnect(context.Background())
	}()

	db := client.Database(dbName)
	taskColl := db.Collection(taskCollName)
	userColl := db.Collection(userCollName)

	// services
	userService := data.NewUserService(userColl)
	taskService := data.NewTaskService(taskColl)

	// controller
	controller := controllers.NewController(userService, taskService)

	// middleware with jwt secret
	authMw := middleware.NewAuthMiddleware(jwtSecret, userService)

	// router
	r := router.SetupRouter(controller, authMw)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	addr := fmt.Sprintf(":%s", port)
	if err := r.Run(addr); err != nil {
		log.Fatalf("failed to run server: %v", err)
	}
}
