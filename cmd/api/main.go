package main

import (
	"context"
	"log"
	"os"
	"time"

	"cloud.google.com/go/storage"
	firebase "firebase.google.com/go/v4"
	"github.com/joho/godotenv"
	"github.com/ramadhantriyant/pockmon/internal/config"
	"github.com/ramadhantriyant/pockmon/internal/scheduler"
)

func main() {
	ctx := context.Background()
	if err := godotenv.Load(); err != nil {
		log.Println("failed to load .env, using environment variables")
	}

	if err := setupGoogleCredentials(); err != nil {
		log.Fatalf("failed to set up Google credentials: %v", err)
	}

	firebaseApp, err := firebase.NewApp(ctx, nil)
	if err != nil {
		log.Fatalf("firebase init error: %v", err)
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

	storageClient, err := storage.NewClient(ctx)
	if err != nil {
		log.Fatalf("failed to create storage client: %v", err)
	}
	defer storageClient.Close()

	appConfig := config.New(db, firebaseApp)
	appConfig.StorageClient = storageClient
	appConfig.StorageBucket = os.Getenv("STORAGE_BUCKET")

	server := createServer(ctx, appConfig, ":8080")

	cronScheduler := scheduler.Start(ctx, db)
	defer cronScheduler.Stop()

	if err := runServer(ctx, server, 5*time.Second); err != nil {
		log.Fatalf("error running server: %v", err)
	}
}
