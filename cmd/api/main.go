package main

import (
	"context"
	"log"
	"os"
	"time"

	firebase "firebase.google.com/go/v4"
	"github.com/joho/godotenv"
	"github.com/ramadhantriyant/pockmon/internal/config"
	"google.golang.org/api/option"
)

func main() {
	ctx := context.Background()
	if err := godotenv.Load(); err != nil {
		log.Println("failed to load .env, using environment variables")
	}

	opt := option.WithCredentialsFile("firebase-pockmon.json")
	firebaseApp, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		log.Fatalf("firebase option error: %v", err)
	}

	dbURL := os.Getenv("DB_URL")
	if err := runMigration(dbURL); err != nil {
		log.Fatalf("run migration failed: %v", err)
	}

	db, err := connectDatabase(ctx, dbURL)
	if err != nil {
		log.Fatalf("failed connecting to database: %v", err)
	}
	defer db.Close()

	appConfig := config.New(db, firebaseApp)
	server := createServer(ctx, appConfig, ":8080")
	if err := runServer(ctx, server, 5*time.Second); err != nil {
		log.Fatalf("error running server: %v", err)
	}
}
